#include <stdlib.h>
#include <string.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct Image {
  int width;
  int height;
  int channels;
  float *data;
} Image;

typedef struct ACESMetadata {
  // Display window (full canvas)
  int display_x;
  int display_y;
  int display_width;
  int display_height;

  // Data window (actual pixel data region)
  int data_x;
  int data_y;
  int data_width;
  int data_height;

  // Pixel aspect ratio (1.0 for square pixels, 2.0 for 2x anamorphic, etc.)
  float pixel_aspect_ratio;

  // Optional timecode (can be NULL or empty string)
  const char *timecode;

  // ACES version string (e.g., "ACES 1.3")
  const char *aces_version;
} ACESMetadata;

typedef enum {
  RAW = 0,
  LINEARISE_SRGB = 1,
  CONVERT_TO_ACESCG = 2,
} ReadImageOptions;

Image *read_image(const char *filename, char **error_msg,
                  ReadImageOptions options);
void free_image(Image *image);

// Returns 0 on success, non-zero error code on failure
// hdr: 1 for HDR, 0 for LDR
int write_image(const char *filename, Image *image, char **error_msg, int hdr);

// Write image in ACEScg color space with metadata (EXR format)
// Returns 0 on success, non-zero error code on failure
int write_image_aces(const char *filename, Image *image, ACESMetadata *metadata,
                     char **error_msg);

#ifdef __cplusplus
}
#endif
