from gunicorn.app.wsgiapp import WSGIApplication
from gunicorn.app.base import BaseApplication


class GunicornServer(WSGIApplication):
    def __init__(self, app, options=None):
        self.application = app
        self.options = options or {}
        super().__init__()

    def load_config(self):
        # #Load default configuration
        # super().load_config()

        # Apply custom configuration
        for key, value in self.options.items():
            self.cfg.set(key.lower(), value)

    def load_wsgiapp(self):
        return self.application


class GunicornBaseServer(BaseApplication):
    def __init__(self, app, options=None):
        self.application = app
        self.options = options or {}
        super().__init__()

    def load_config(self):
        # #Load default configuration
        # super().load_config()

        # Apply custom configuration
        for key, value in self.options.items():
            self.cfg.set(key.lower(), value)

    def load(self):
        return self.application