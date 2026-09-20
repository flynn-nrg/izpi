#include "texturesys_wrapper.h"
#include <OpenImageIO/texture.h>
#include <cstdlib>
#include <cstring>

using namespace OIIO;

struct TextureSystemHandle {
  std::shared_ptr<TextureSystem> ts;
};

namespace {

Tex::Wrap to_oiio_wrap(TextureWrapMode w) {
  switch (w) {
  case WRAP_BLACK:
    return Tex::Wrap::Black;
  case WRAP_CLAMP:
    return Tex::Wrap::Clamp;
  case WRAP_PERIODIC:
    return Tex::Wrap::Periodic;
  case WRAP_MIRROR:
    return Tex::Wrap::Mirror;
  case WRAP_DEFAULT:
  default:
    return Tex::Wrap::Default;
  }
}

TextureOpt to_oiio_opt(const TextureLookupOptions *opts) {
  TextureOpt opt;
  opt.firstchannel = opts->firstchannel;
  opt.subimage = opts->subimage;
  opt.swrap = to_oiio_wrap(opts->swrap);
  opt.twrap = to_oiio_wrap(opts->twrap);
  opt.rwrap = to_oiio_wrap(opts->rwrap);
  opt.sblur = opts->sblur;
  opt.tblur = opts->tblur;
  opt.rblur = opts->rblur;
  opt.swidth = opts->swidth;
  opt.twidth = opts->twidth;
  opt.rwidth = opts->rwidth;
  opt.fill = opts->fill;
  return opt;
}

} // namespace

TextureSystemHandle *texturesystem_create(char **error_msg) {
  auto ts = TextureSystem::create(/*shared=*/true);
  if (!ts) {
    if (error_msg)
      *error_msg = strdup("Could not create OIIO TextureSystem");
    return nullptr;
  }
  auto *handle = new TextureSystemHandle();
  handle->ts = ts;
  return handle;
}

void texturesystem_destroy(TextureSystemHandle *handle) { delete handle; }

int texturesystem_texture(TextureSystemHandle *handle, const char *filename,
                          const TextureLookupOptions *opts, float s, float t,
                          float dsdx, float dtdx, float dsdy, float dtdy,
                          int nchannels, float *result, char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  TextureOpt opt = to_oiio_opt(opts);
  bool ok = handle->ts->texture(ustring(filename), opt, s, t, dsdx, dtdx, dsdy,
                                dtdy, nchannels, result);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_texture3d(TextureSystemHandle *handle, const char *filename,
                            const TextureLookupOptions *opts, const float *p,
                            const float *dpdx, const float *dpdy,
                            const float *dpdz, int nchannels, float *result,
                            char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  TextureOpt opt = to_oiio_opt(opts);
  bool ok = handle->ts->texture3d(
      ustring(filename), opt, V3fParam(p[0], p[1], p[2]),
      V3fParam(dpdx[0], dpdx[1], dpdx[2]), V3fParam(dpdy[0], dpdy[1], dpdy[2]),
      V3fParam(dpdz[0], dpdz[1], dpdz[2]), nchannels, result);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_environment(TextureSystemHandle *handle,
                              const char *filename,
                              const TextureLookupOptions *opts,
                              const float *r, const float *drdx,
                              const float *drdy, int nchannels, float *result,
                              char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  TextureOpt opt = to_oiio_opt(opts);
  bool ok = handle->ts->environment(
      ustring(filename), opt, V3fParam(r[0], r[1], r[2]),
      V3fParam(drdx[0], drdx[1], drdx[2]), V3fParam(drdy[0], drdy[1], drdy[2]),
      nchannels, result);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_get_texture_info_int(TextureSystemHandle *handle,
                                       const char *filename,
                                       const char *dataname, int *out,
                                       char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  bool ok = handle->ts->get_texture_info(ustring(filename), 0,
                                         ustring(dataname), TypeDesc::INT, out);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_get_texture_info_float(TextureSystemHandle *handle,
                                         const char *filename,
                                         const char *dataname, float *out,
                                         char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  bool ok = handle->ts->get_texture_info(
      ustring(filename), 0, ustring(dataname), TypeDesc::FLOAT, out);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

namespace {

// attribute()/getattribute() report failure by returning false without
// necessarily populating geterror() (an unrecognized attribute name/type is
// not logged as an error), so fall back to a fixed message.
char *strdup_error_or(const std::string &err, const char *fallback) {
  return strdup(err.empty() ? fallback : err.c_str());
}

} // namespace

int texturesystem_attribute_int(TextureSystemHandle *handle, const char *name,
                                int value, char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  if (!handle->ts->attribute(name, value)) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "unrecognized int attribute");
    return 1;
  }
  return 0;
}

int texturesystem_attribute_float(TextureSystemHandle *handle,
                                  const char *name, float value,
                                  char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  if (!handle->ts->attribute(name, value)) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "unrecognized float attribute");
    return 1;
  }
  return 0;
}

int texturesystem_attribute_string(TextureSystemHandle *handle,
                                   const char *name, const char *value,
                                   char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  if (!handle->ts->attribute(name, string_view(value))) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "unrecognized string attribute");
    return 1;
  }
  return 0;
}

int texturesystem_getattribute_int(TextureSystemHandle *handle,
                                   const char *name, int *out,
                                   char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  if (!handle->ts->getattribute(name, *out)) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "unrecognized int attribute");
    return 1;
  }
  return 0;
}

int texturesystem_getattribute_float(TextureSystemHandle *handle,
                                     const char *name, float *out,
                                     char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  if (!handle->ts->getattribute(name, *out)) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "unrecognized float attribute");
    return 1;
  }
  return 0;
}

int texturesystem_getattribute_string(TextureSystemHandle *handle,
                                      const char *name, char **out,
                                      char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return 1;
  }
  std::string value;
  if (!handle->ts->getattribute(name, value)) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "unrecognized string attribute");
    return 1;
  }
  *out = strdup(value.c_str());
  return 0;
}

namespace {
inline TextureSystem::Perthread *to_oiio_perthread(PerThreadInfo *info) {
  return reinterpret_cast<TextureSystem::Perthread *>(info);
}
inline TextureSystem::TextureHandle *to_oiio_texhandle(TextureHandle *th) {
  return reinterpret_cast<TextureSystem::TextureHandle *>(th);
}
} // namespace

PerThreadInfo *texturesystem_create_thread_info(TextureSystemHandle *handle) {
  if (!handle || !handle->ts)
    return nullptr;
  return reinterpret_cast<PerThreadInfo *>(handle->ts->create_thread_info());
}

void texturesystem_destroy_thread_info(TextureSystemHandle *handle,
                                       PerThreadInfo *info) {
  if (!handle || !handle->ts || !info)
    return;
  handle->ts->destroy_thread_info(to_oiio_perthread(info));
}

TextureHandle *texturesystem_get_texture_handle(TextureSystemHandle *handle,
                                                const char *filename,
                                                PerThreadInfo *thread_info,
                                                char **error_msg) {
  if (!handle || !handle->ts) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem handle");
    return nullptr;
  }
  auto *th = handle->ts->get_texture_handle(ustring(filename),
                                            to_oiio_perthread(thread_info));
  if (!th || !handle->ts->good(th)) {
    if (error_msg)
      *error_msg = strdup_error_or(handle->ts->geterror(),
                                   "could not resolve texture handle");
    return nullptr;
  }
  return reinterpret_cast<TextureHandle *>(th);
}

int texturesystem_texture_handle(TextureSystemHandle *handle,
                                 TextureHandle *th, PerThreadInfo *thread_info,
                                 const TextureLookupOptions *opts, float s,
                                 float t, float dsdx, float dtdx, float dsdy,
                                 float dtdy, int nchannels, float *result,
                                 char **error_msg) {
  if (!handle || !handle->ts || !th) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem/TextureHandle");
    return 1;
  }
  TextureOpt opt = to_oiio_opt(opts);
  bool ok = handle->ts->texture(to_oiio_texhandle(th),
                                to_oiio_perthread(thread_info), opt, s, t,
                                dsdx, dtdx, dsdy, dtdy, nchannels, result);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_texture3d_handle(TextureSystemHandle *handle,
                                   TextureHandle *th,
                                   PerThreadInfo *thread_info,
                                   const TextureLookupOptions *opts,
                                   const float *p, const float *dpdx,
                                   const float *dpdy, const float *dpdz,
                                   int nchannels, float *result,
                                   char **error_msg) {
  if (!handle || !handle->ts || !th) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem/TextureHandle");
    return 1;
  }
  TextureOpt opt = to_oiio_opt(opts);
  bool ok = handle->ts->texture3d(
      to_oiio_texhandle(th), to_oiio_perthread(thread_info), opt,
      V3fParam(p[0], p[1], p[2]), V3fParam(dpdx[0], dpdx[1], dpdx[2]),
      V3fParam(dpdy[0], dpdy[1], dpdy[2]), V3fParam(dpdz[0], dpdz[1], dpdz[2]),
      nchannels, result);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_environment_handle(TextureSystemHandle *handle,
                                     TextureHandle *th,
                                     PerThreadInfo *thread_info,
                                     const TextureLookupOptions *opts,
                                     const float *r, const float *drdx,
                                     const float *drdy, int nchannels,
                                     float *result, char **error_msg) {
  if (!handle || !handle->ts || !th) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem/TextureHandle");
    return 1;
  }
  TextureOpt opt = to_oiio_opt(opts);
  bool ok = handle->ts->environment(
      to_oiio_texhandle(th), to_oiio_perthread(thread_info), opt,
      V3fParam(r[0], r[1], r[2]), V3fParam(drdx[0], drdx[1], drdx[2]),
      V3fParam(drdy[0], drdy[1], drdy[2]), nchannels, result);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_get_texture_info_handle_int(TextureSystemHandle *handle,
                                              TextureHandle *th,
                                              PerThreadInfo *thread_info,
                                              const char *dataname, int *out,
                                              char **error_msg) {
  if (!handle || !handle->ts || !th) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem/TextureHandle");
    return 1;
  }
  bool ok = handle->ts->get_texture_info(to_oiio_texhandle(th),
                                         to_oiio_perthread(thread_info), 0,
                                         ustring(dataname), TypeDesc::INT,
                                         out);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}

int texturesystem_get_texture_info_handle_float(TextureSystemHandle *handle,
                                                TextureHandle *th,
                                                PerThreadInfo *thread_info,
                                                const char *dataname,
                                                float *out,
                                                char **error_msg) {
  if (!handle || !handle->ts || !th) {
    if (error_msg)
      *error_msg = strdup("Invalid TextureSystem/TextureHandle");
    return 1;
  }
  bool ok = handle->ts->get_texture_info(to_oiio_texhandle(th),
                                         to_oiio_perthread(thread_info), 0,
                                         ustring(dataname), TypeDesc::FLOAT,
                                         out);
  if (!ok) {
    if (error_msg)
      *error_msg = strdup(handle->ts->geterror().c_str());
    return 1;
  }
  return 0;
}
