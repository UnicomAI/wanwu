from gunicorn.app.wsgiapp import WSGIApplication


class GunicornServer(WSGIApplication):
    def __init__(self, app, options=None):
        self.application = app
        self.options = options or {}
        super().__init__()

    def load_config(self):
        # # 加载默认配置 [EN] #Load default configuration
        # super().load_config()

        # 应用自定义配置 [EN] Apply custom configuration
        for key, value in self.options.items():
            self.cfg.set(key.lower(), value)

    def load_wsgiapp(self):
        return self.application