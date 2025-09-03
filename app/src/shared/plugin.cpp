#include "shared/plugin.h"
#include <map>
#include <string>
#include <memory>
#include <mutex>
#include <atomic>
#include <functional>
#include <fstream>

#include "include/cef_app.h"
#include "include/cef_client.h"
#include "include/cef_browser.h"
#include "include/cef_request_context.h"
#include "include/cef_request.h"
#include "include/cef_frame.h"
#include "include/cef_parser.h"

static std::mutex g_logMutex;

inline void LogToFile(const std::string &msg) {
#ifdef _DEBUG
    std::lock_guard<std::mutex> lock(g_logMutex);
    std::ofstream ofs("cef_plugin.log", std::ios::app);
    ofs << msg << std::endl;
#endif
}

class MinimalApp : public CefApp, public CefBrowserProcessHandler,
                   public CefRenderProcessHandler {
public:
    explicit MinimalApp(const std::string &user_agent) : user_agent_app_(user_agent) {}

    CefRefPtr<CefBrowserProcessHandler> GetBrowserProcessHandler() override {
        return this;
    }

    CefRefPtr<CefRenderProcessHandler> GetRenderProcessHandler() override {
        return this;
    }

    void OnContextCreated(CefRefPtr<CefBrowser> browser,
                          CefRefPtr<CefFrame> frame,
                          CefRefPtr<CefV8Context> context) override {
//        if (frame->IsMain() && !user_agent_app_.empty()) {
//            std::string js_code = "Object.defineProperty(navigator, 'userAgent', {"
//                                  "  get: function() { return '" + user_agent_app_ + "'; },"
//                                                                                     "  configurable: true"
//                                                                                     "});";
//            frame->ExecuteJavaScript(js_code, frame->GetURL(), 0);
//            LogToFile("Injected JS in OnContextCreated to set navigator.userAgent: " + user_agent_app_);
//        }
    }

    void OnBeforeCommandLineProcessing(const CefString &process_type,
                                       CefRefPtr<CefCommandLine> cmd) override {
        cmd->AppendSwitch("disable-gpu");
        cmd->AppendSwitch("disable-gpu-compositing");
        cmd->AppendSwitch("disable-software-rasterizer");
        cmd->AppendSwitch("no-sandbox");
        cmd->AppendSwitch("disable-renderer-process-sharing");
        LogToFile("Applied command line switches for process: " + process_type.ToString());
    }

IMPLEMENT_REFCOUNTING(MinimalApp);

private:
    std::string user_agent_app_;
};

class PerRequestRRH : public CefResourceRequestHandler {
public:
    PerRequestRRH(const std::string &user_agent,
                  const std::map<std::string, std::string> *extra_headers)
            : user_agent_req_(user_agent), extra_headers_(extra_headers) {}

    cef_return_value_t OnBeforeResourceLoad(CefRefPtr<CefBrowser> browser,
                                            CefRefPtr<CefFrame> frame,
                                            CefRefPtr<CefRequest> request,
                                            CefRefPtr<CefCallback> callback) override {
        if (!user_agent_req_.empty()) {
            request->SetHeaderByName("User-Agent", user_agent_req_, true);
            LogToFile("Set User-Agent: " + user_agent_req_ + " for URL: " + std::string(request->GetURL()));
        }
        if (extra_headers_) {
            for (const auto &kv: *extra_headers_) {
                request->SetHeaderByName(kv.first, kv.second, true);
            }
        }
        return RV_CONTINUE;
    }

private:
    std::string user_agent_req_;
    const std::map<std::string, std::string> *extra_headers_;

IMPLEMENT_REFCOUNTING(PerRequestRRH);
    DISALLOW_COPY_AND_ASSIGN(PerRequestRRH);
};

class BrowserClient : public CefClient,
                      public CefLifeSpanHandler,
                      public CefRequestHandler,
                      public CefLoadHandler,
                      public CefRenderProcessHandler {
public:
    BrowserClient(std::function<void(CefRefPtr<CefBrowser>)> on_created,
                  const std::string &user_agent,
                  const std::map<std::string, std::string> *extra_headers,
                  const std::string &js_patch)
            : on_created_(std::move(on_created)),
              rrh_(new PerRequestRRH(user_agent, extra_headers)),
              user_agent_cli_(user_agent),
              js_patch_(js_patch) {}

    CefRefPtr<CefLifeSpanHandler> GetLifeSpanHandler() override { return this; }

    CefRefPtr<CefRequestHandler> GetRequestHandler() override { return this; }

    CefRefPtr<CefLoadHandler> GetLoadHandler() override { return this; }

    void OnAfterCreated(CefRefPtr<CefBrowser> browser) override {
        if (on_created_) on_created_(browser);
        LogToFile("Browser created with ID: " + std::to_string(browser->GetIdentifier()));
    }

    void OnContextCreated(CefRefPtr<CefBrowser> browser,
                          CefRefPtr<CefFrame> frame,
                          CefRefPtr<CefV8Context> context) override {
        if (frame->IsMain() && !user_agent_cli_.empty()) {
            std::string js_code = "Object.defineProperty(navigator, 'userAgent', {"
                                  "  get: function() { return '" + user_agent_cli_ + "'; },"
                                                                                     "  configurable: true"
                                                                                     "});"
                                                                                     "Object.defineProperty(navigator, 'plugins', {"
                                                                                     "  get: function() { return []; },"
                                                                                     "  configurable: true"
                                                                                     "});"
                                                                                     "Object.defineProperty(window, 'screen', {"
                                                                                     "  get: function() { return { width: 1920, height: 1080 }; },"
                                                                                     "  configurable: true"
                                                                                     "});";
            frame->ExecuteJavaScript(js_code, frame->GetURL(), 0);
            LogToFile("Injected JS in OnContextCreated to set navigator.userAgent: " + user_agent_cli_ +
                      " for browser ID: " + std::to_string(browser->GetIdentifier()));
        }
        if (frame->IsMain() && !js_patch_.empty()) {
            frame->ExecuteJavaScript(js_patch_, frame->GetURL(), 0);
            LogToFile("Injected custom JS patch in OnContextCreated: " + js_patch_ +
                      " for browser ID: " + std::to_string(browser->GetIdentifier()));
        }
    }

    void OnLoadStart(CefRefPtr<CefBrowser> browser,
                     CefRefPtr<CefFrame> frame,
                     TransitionType transition_type) override {
        if (frame->IsMain() && !user_agent_cli_.empty()) {
            std::string js_code = "try {"
                                  "  Object.defineProperty(navigator, 'userAgent', {"
                                  "    get: function() { return '" + user_agent_cli_ + "'; },"
                                                                                       "    configurable: true"
                                                                                       "  });"
                                                                                       "  console.log('Set navigator.userAgent to: " +
                                  user_agent_cli_ + " in OnLoadStart');"
                                                    "} catch (e) {"
                                                    "  console.error('Failed to set navigator.userAgent in OnLoadStart: ' + e.message);"
                                                    "}";
            frame->ExecuteJavaScript(js_code, frame->GetURL(), 0);
            LogToFile("Injected JS in OnLoadStart to set navigator.userAgent: " + user_agent_cli_ +
                      " for browser ID: " + std::to_string(browser->GetIdentifier()) +
                      " URL: " + std::string(frame->GetURL()));
        }
    }

    void OnLoadEnd(CefRefPtr<CefBrowser> browser,
                   CefRefPtr<CefFrame> frame,
                   int httpStatusCode) override {
        if (frame->IsMain() && !js_patch_.empty()) {
            frame->ExecuteJavaScript(js_patch_, frame->GetURL(), 0);
            LogToFile("Injected custom JS patch in OnLoadEnd: " + js_patch_ +
                      " for browser ID: " + std::to_string(browser->GetIdentifier()));
        }
    }

    CefRefPtr<CefResourceRequestHandler> GetResourceRequestHandler(
            CefRefPtr<CefBrowser>, CefRefPtr<CefFrame>, CefRefPtr<CefRequest>,
            bool, bool, const CefString &, bool &) override {
        return rrh_;
    }

private:
    std::function<void(CefRefPtr<CefBrowser>)> on_created_;
    CefRefPtr<PerRequestRRH> rrh_;
    std::string user_agent_cli_;
    std::string js_patch_;

IMPLEMENT_REFCOUNTING(BrowserClient);
    DISALLOW_COPY_AND_ASSIGN(BrowserClient);
};

class BrowserInstanceImpl : public std::enable_shared_from_this<BrowserInstanceImpl> {
public:
    BrowserInstanceImpl(const std::string &userDataDir,
                        const std::string &userAgent,
                        const std::string &proxyServer,
                        const std::string &bypassList,
                        const std::string &jsPatch)
            : user_data_dir_(userDataDir),
              user_agent_(userAgent),
              proxy_server_(proxyServer),
              bypass_list_(bypassList),
              js_patch_(jsPatch) {
        InitCefInstance();
        if (cef_initialized_) {
            CreateBrowser();
        }
    }

    ~BrowserInstanceImpl() {
        if (browser_) {
            browser_->GetHost()->CloseBrowser(true);
        }
    }

    bool IsInitialized() const {
        return cef_initialized_;
    }

    void LoadURL(const std::string &url) {
        if (browser_) browser_->GetMainFrame()->LoadURL(url);
        else pending_url_ = url;

        if (browser_) {
            browser_->GetMainFrame()->LoadURL(url);
            LogToFile("Loading URL: " + url + " for user-agent: " + user_agent_);
        } else {
            pending_url_ = url;
            LogToFile("Pending URL: " + url + " for user-agent: " + user_agent_);
        }
    }

    void EvalJS(const std::string &code) {
        if (browser_) {
            auto frame = browser_->GetMainFrame();
            frame->ExecuteJavaScript(code, frame->GetURL(), 0);
        } else {
            pending_js_ = code;
        }
    }

    void SetHeader(const std::string &name, const std::string &value) {
        std::lock_guard<std::mutex> lock(headers_mutex_);
        extra_headers_[name] = value;
    }

    void SetProxy(const std::string &server, const std::string &bypass) {
        proxy_server_ = server;
        bypass_list_ = bypass;
        ApplyProxy();
    }

private:
    std::string user_data_dir_;
    std::string user_agent_;
    std::string proxy_server_;
    std::string bypass_list_;
    std::string js_patch_;

    std::string pending_url_;
    std::string pending_js_;

    CefRefPtr<CefBrowser> browser_;
    CefRefPtr<BrowserClient> client_;
    CefRefPtr<CefRequestContext> request_context_;
    bool cef_initialized_ = false;

    std::mutex headers_mutex_;
    std::map<std::string, std::string> extra_headers_;

    void InitCefInstance() {
        CefMainArgs main_args(GetModuleHandle(nullptr));
        CefRefPtr<MinimalApp> app = new MinimalApp(user_agent_);

        int exit_code = CefExecuteProcess(main_args, app.get(), nullptr);
        if (exit_code >= 0) {
            ::exit(exit_code);
        }

        CefSettings settings;
        settings.no_sandbox = true;
        settings.windowless_rendering_enabled = false;
        if (!user_data_dir_.empty()) {
            CefString(&settings.root_cache_path) = user_data_dir_;
            CefString(&settings.cache_path) = user_data_dir_;
        }
        CefString(&settings.log_file) = "cef_debug.log";
        settings.log_severity = LOGSEVERITY_VERBOSE;

        cef_initialized_ = CefInitialize(main_args, settings, app.get(), nullptr);
        if (!cef_initialized_) {
            LogToFile("CEF instance initialization failed.");
        } else {
            LogToFile("CEF instance initialized.");
        }
    }

    void CreateBrowser() {
        CefRequestContextSettings ctx_settings;
        if (!user_data_dir_.empty()) {
            CefString(&ctx_settings.cache_path) = user_data_dir_;
        }
        request_context_ = CefRequestContext::CreateContext(ctx_settings, nullptr);

        client_ = new BrowserClient(
                [this](CefRefPtr<CefBrowser> b) {
                    browser_ = b;
                    ApplyProxy();
                    if (!pending_url_.empty())
                        browser_->GetMainFrame()->LoadURL(pending_url_);
                    LogToFile("Loaded pending URL: " + pending_url_ + " for user-agent: " + user_agent_);
                    if (!pending_js_.empty()) {
                        auto f = browser_->GetMainFrame();
                        f->ExecuteJavaScript(pending_js_, f->GetURL(), 0);
                        LogToFile("Executed pending JS: " + pending_js_ + " for user-agent: " + user_agent_);
                    }
                },
                user_agent_,
                &extra_headers_,
                js_patch_
        );

        CefWindowInfo window_info;
        window_info.SetAsPopup(nullptr, "CEF Browser");

        CefBrowserSettings browser_settings;
        CefBrowserHost::CreateBrowser(window_info, client_, "about:blank",
                                      browser_settings, nullptr, request_context_);

        LogToFile("Browser creation requested for user-agent: " + user_agent_);
    }

    void ApplyProxy() {
        if (!request_context_) return;
        if (proxy_server_.empty()) return;

        CefRefPtr<CefDictionaryValue> proxy_dict = CefDictionaryValue::Create();
        proxy_dict->SetString("mode", "fixed_servers");
        proxy_dict->SetString("server", proxy_server_);
        if (!bypass_list_.empty()) {
            proxy_dict->SetString("bypass_list", bypass_list_);
        }

        CefRefPtr<CefValue> proxy_value = CefValue::Create();
        proxy_value->SetDictionary(proxy_dict);

        CefString error;
        const bool ok = request_context_->SetPreference("proxy", proxy_value, error);
        if (!ok) {
            LogToFile("Proxy SetPreference failed: " + std::string(error.ToString()));
        } else {
            LogToFile("Proxy applied: " + proxy_server_ + " bypass: " + bypass_list_);
        }
    }
};

static std::map<PluginInstanceHandle, std::shared_ptr<BrowserInstanceImpl>> g_instances;
static std::mutex g_instancesMutex;
static std::atomic<PluginInstanceHandle> g_nextHandle{1};

PluginInstanceHandle PluginCreateInstance(const char *userDataDir,
                                          const char *userAgent,
                                          const char *proxyServer,
                                          const char *bypassList,
                                          const char *jsFingerprintPatch,
                                          int *error_code) {
    std::lock_guard<std::mutex> lock(g_instancesMutex);
    auto inst = std::make_shared<BrowserInstanceImpl>(
            userDataDir ? userDataDir : "",
            userAgent ? userAgent : "",
            proxyServer ? proxyServer : "",
            bypassList ? bypassList : "",
            jsFingerprintPatch ? jsFingerprintPatch : ""
    );
    if (!inst->IsInitialized()) {
        *error_code = 1;
        LogToFile("Failed to create instance for user-agent: " + std::string(userAgent ? userAgent : ""));
        return 0;
    }
    PluginInstanceHandle handle = g_nextHandle++;
    g_instances[handle] = inst;
    *error_code = 0;
    LogToFile("Created instance with handle: " + std::to_string(handle) +
              " for user-agent: " + std::string(userAgent ? userAgent : ""));
    return handle;
}

void PluginDestroyInstance(PluginInstanceHandle handle) {
    std::lock_guard<std::mutex> lock(g_instancesMutex);
    auto it = g_instances.find(handle);
    if (it != g_instances.end()) {
        LogToFile("Destroying instance with handle: " + std::to_string(handle));
        g_instances.erase(it);
    }
}

void PluginLoadURL(PluginInstanceHandle handle, const char *url) {
    std::lock_guard<std::mutex> lock(g_instancesMutex);
    auto it = g_instances.find(handle);
    if (it != g_instances.end()) {
        it->second->LoadURL(url ? url : "");
        LogToFile("Loading URL: " + std::string(url ? url : "") + " for handle: " + std::to_string(handle));
    }
}

void PluginEvalJS(PluginInstanceHandle handle, const char *jsCode) {
    std::lock_guard<std::mutex> lock(g_instancesMutex);
    auto it = g_instances.find(handle);
    if (it != g_instances.end()) {
        it->second->EvalJS(jsCode ? jsCode : "");
        LogToFile("Executing JS: " + std::string(jsCode ? jsCode : "") + " for handle: " + std::to_string(handle));
    }
}

void PluginSetHeader(PluginInstanceHandle handle, const char *name, const char *value) {
    std::lock_guard<std::mutex> lock(g_instancesMutex);
    auto it = g_instances.find(handle);
    if (it != g_instances.end()) {
        it->second->SetHeader(name ? name : "", value ? value : "");
        LogToFile("Set header: " + std::string(name ? name : "") + "=" + std::string(value ? value : "") +
                  " for handle: " + std::to_string(handle));
    }
}

void PluginSetProxy(PluginInstanceHandle handle, const char *proxyServer, const char *bypassList) {
    std::lock_guard<std::mutex> lock(g_instancesMutex);
    auto it = g_instances.find(handle);
    if (it != g_instances.end()) {
        it->second->SetProxy(proxyServer ? proxyServer : "", bypassList ? bypassList : "");
        LogToFile("Set proxy: " + std::string(proxyServer ? proxyServer : "") +
                  " bypass: " + std::string(bypassList ? bypassList : "") +
                  " for handle: " + std::to_string(handle));
    }
}