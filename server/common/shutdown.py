import logging
import signal

class ShutdownHandler:
    def __init__(self, server):
        self._server = server
        signal.signal(signal.SIGTERM, self.exit_gracefully)
        signal.signal(signal.SIGINT, self.exit_gracefully)

    def exit_gracefully(self, signum, frame):
        logging.info("action: exit | result: success")
        self._server.stop()

