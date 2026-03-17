import logging
import signal

class ShutdownHandler:
    def __init__(self, server):
        self._server = server
        signal.signal(signal.SIGTERM, self.exit_gracefully)
        signal.signal(signal.SIGINT, self.exit_gracefully)

    def exit_gracefully(self):
        logging.info("action: shutdown_signal_received | result: success")
        self._server.stop()

