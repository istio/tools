from pickle import Pickler, Unpickler
from threading import Thread
from time import sleep
from modules.sampler import Sampler
import os
import tempfile


class Collector:
    def __init__(self, file, sampler=Sampler(), sample_interval=1.0):
        self.file = file
        self.sample_interval = sample_interval
        self.thread = Thread(target=self.work)
        self.sampler = sampler

    def work(self):
        while self.running:
            self.pickler.dump(self.sampler.get_snapshot())
            # We clear this so the pickler won't remember which objects
            # it has already seen. This allows us to restore flattened
            # process structured, thereby serializing a flattened version
            # into yaml.
            self.pickler.clear_memo()
            sleep(self.sample_interval)

    def start(self):
        self.running = True
        self.pickler = Pickler(self.file)
        self.thread.start()

    def stop(self):
        self.running = False
        self.thread.join()
        self.file.close()

    def read_dump(self, file):
        unpickler = Unpickler(file)
        done = False
        while not done:
            try:
                yield unpickler.load()
            except EOFError:
                done = True
