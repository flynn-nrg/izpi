#include "oiio_wrapper.h"
#include <OpenImageIO/imagebuf.h>
#include <OpenImageIO/imagebufalgo.h>
#include <OpenImageIO/imageio.h>

using namespace OIIO;

Image *read_image_raw(const char *filename, char **error_msg) {
  Image *image = new Image();

  auto inp = ImageInput::open(filename);
  if (!inp) {
    *error_msg = strdup("Could not open image file");
    return nullptr;
  }

  const ImageSpec &spec = inp->spec();
  int xres = spec.width;
  int yres = spec.height;
  int nchannels = spec.nchannels;

  image->width = xres;
  image->height = yres;
  image->channels = nchannels;

  image->data = new float[xres * yres * nchannels];

  if (!inp->read_image(0, 0, 0, nchannels, TypeDesc::FLOAT, image->data)) {
    *error_msg = strdup(inp->geterror().c_str());
    delete[] image->data;
    delete image;
    inp->close();
    return nullptr;
  }

  inp->close();
  return image;
}

Image *read_image_with_conversion(const char *filename, char **error_msg,
                                  const char *destination_color_space) {
  Image *image = new Image();

  // Read image into an ImageBuf
  ImageBuf src(filename);
  if (src.has_error()) {
    *error_msg = strdup(src.geterror().c_str());
    delete image;
    return nullptr;
  }

  const ImageSpec &spec = src.spec();
  int xres = spec.width;
  int yres = spec.height;
  int nchannels = spec.nchannels;

  // Perform color conversion from sRGB to destination_color_space using OCIO
  // Using "sRGB Encoded Rec.709 (sRGB)" which is standard in OCIO 2.x built-in
  // configs
  ImageBuf dst;
  if (!ImageBufAlgo::colorconvert(dst, src, "sRGB Encoded Rec.709 (sRGB)",
                                  destination_color_space)) {
    *error_msg = strdup(dst.geterror().c_str());
    delete image;
    return nullptr;
  }

  image->width = xres;
  image->height = yres;
  image->channels = nchannels;

  // Allocate and copy the converted data
  image->data = new float[xres * yres * nchannels];
  if (!dst.get_pixels(ROI::All(), TypeDesc::FLOAT, image->data)) {
    *error_msg = strdup(dst.geterror().c_str());
    delete[] image->data;
    delete image;
    return nullptr;
  }

  return image;
}

Image *read_image(const char *filename, char **error_msg,
                  ReadImageOptions options) {
  switch (options) {
  case LINEARISE_SRGB:
    return read_image_with_conversion(filename, error_msg,
                                      "Linear Rec.709 (sRGB)");
  case CONVERT_TO_ACESCG:
    return read_image_with_conversion(filename, error_msg, "ACEScg");
  case RAW:
    return read_image_raw(filename, error_msg);
  default:
    *error_msg = strdup("Invalid read image options");
    return nullptr;
  }
}

void free_image(Image *image) {
  delete[] image->data;
  delete image;
}

int write_image(const char *filename, Image *image, char **error_msg, int hdr) {
  std::unique_ptr<ImageOutput> out = ImageOutput::create(filename);
  if (!out) {
    *error_msg = strdup("Could not create ImageOutput");
    return 1;
  }

  ImageSpec spec;
  spec.width = image->width;
  spec.height = image->height;
  spec.nchannels = image->channels;
  spec.format = hdr ? TypeDesc::FLOAT : TypeDesc::UINT8;
  spec.channelnames = {"R", "G", "B", "A"};

  if (!out->open(filename, spec)) {
    *error_msg = strdup(out->geterror().c_str());
    return 1;
  }

  if (hdr) {
    // For HDR formats, write the float data directly
    if (!out->write_image(TypeDesc::FLOAT, image->data)) {
      *error_msg = strdup(out->geterror().c_str());
      out->close();
      return 1;
    }
  } else {
    // For LDR formats, convert float to UINT8
    size_t num_pixels = image->width * image->height * image->channels;
    std::vector<uint8_t> uint8_data(num_pixels);
    for (size_t i = 0; i < num_pixels; ++i) {
      uint8_data[i] = static_cast<uint8_t>(
          std::min(255.0f, std::max(0.0f, image->data[i] * 255.0f)));
    }
    if (!out->write_image(TypeDesc::UINT8, uint8_data.data())) {
      *error_msg = strdup(out->geterror().c_str());
      out->close();
      return 1;
    }
  }

  out->close();
  return 0;
}

// Write image in ACEScg color space with metadata
int write_image_aces(const char *filename, Image *image, ACESMetadata *metadata,
                     char **error_msg) {
  std::unique_ptr<ImageOutput> out = ImageOutput::create(filename);
  if (!out) {
    *error_msg = strdup("Could not create ImageOutput for ACES");
    return 1;
  }

  ImageSpec spec;
  spec.width = image->width;
  spec.height = image->height;
  spec.nchannels = image->channels;
  spec.format = TypeDesc::FLOAT; // Always use float for ACES/EXR
  spec.channelnames = {"R", "G", "B", "A"};

  // Set color space to ACEScg
  spec.attribute("oiio:ColorSpace", "ACEScg");

  // Set ACEScg (AP1) chromaticities
  float chromaticities[8] = {
      0.713f,   0.300f,  // Red primary
      0.165f,   0.830f,  // Green primary
      0.128f,   0.044f,  // Blue primary
      0.32168f, 0.33767f // White point (D60)
  };
  spec.attribute("chromaticities", TypeDesc(TypeDesc::FLOAT, 8),
                 chromaticities);

  // Set display window (full canvas) and pixel aspect ratio
  spec.attribute("PixelAspectRatio", metadata->pixel_aspect_ratio);
  spec.full_x = metadata->display_x;
  spec.full_y = metadata->display_y;
  spec.full_width = metadata->display_width;
  spec.full_height = metadata->display_height;

  // Set data window (actual pixel data region)
  spec.x = metadata->data_x;
  spec.y = metadata->data_y;
  spec.width = metadata->data_width;
  spec.height = metadata->data_height;

  // Set timecode if provided
  if (metadata->timecode && strlen(metadata->timecode) > 0) {
    spec.attribute("smpte:TimeCode", metadata->timecode);
  }

  // Set ACES version as custom attribute
  if (metadata->aces_version && strlen(metadata->aces_version) > 0) {
    spec.attribute("aces:version", metadata->aces_version);
  }

  // Add standard ACES metadata
  spec.attribute("compression", "zip"); // Standard compression for ACES EXR
  spec.attribute("openexr:lineOrder", "increasing");

  if (!out->open(filename, spec)) {
    *error_msg = strdup(out->geterror().c_str());
    return 1;
  }

  // Write float data directly (no conversion needed for EXR)
  if (!out->write_image(TypeDesc::FLOAT, image->data)) {
    *error_msg = strdup(out->geterror().c_str());
    out->close();
    return 1;
  }

  out->close();
  return 0;
}
