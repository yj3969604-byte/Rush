#ifndef PLUGIN_H
#define PLUGIN_H

#ifdef _WIN32
#ifdef BUILDING_PLUGIN_DLL
#define PLUGIN_API __declspec(dllexport)
#else
#define PLUGIN_API __declspec(dllimport)
#endif
#else
#define PLUGIN_API
#endif

#include <cstdint>

#ifdef __cplusplus
extern "C" {
#endif

typedef int32_t PluginInstanceHandle;

PLUGIN_API PluginInstanceHandle
PluginCreateInstance(const char *userDataDir, const char *userAgent, const char *proxyServer, const char *bypassList,
                     const char *jsFingerprintPatch, int *error_code);

PLUGIN_API void PluginDestroyInstance(PluginInstanceHandle handle);
PLUGIN_API void PluginLoadURL(PluginInstanceHandle handle, const char *url);
PLUGIN_API void PluginEvalJS(PluginInstanceHandle handle, const char *jsCode);
PLUGIN_API void PluginSetHeader(PluginInstanceHandle handle, const char *name, const char *value);
PLUGIN_API void PluginSetProxy(PluginInstanceHandle handle, const char *proxyServer, const char *bypassList);
PLUGIN_API void PluginClearProxy(PluginInstanceHandle handle);

#ifdef __cplusplus
}
#endif
#endif // PLUGIN_H