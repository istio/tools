#!/usr/bin/env python3

import yaml
from enum import Enum
from yaml import CDumper as CDumper

# We need to massage yaml serialization a bit we we'll output named tuples in a nice way.
# Also, we use CDumper, which is a little faster then the stock dumper.


def to_yaml(o):
    def setup_yaml_formatting():
        def tuple_formatter(self, data):
            if hasattr(data, '_asdict'):
                return self.represent_dict(data._asdict())
            return self.represent_list(data)

        def enum_formatter(self, data):
            return self.represent_data(repr(data))

        yaml.Dumper.yaml_multi_representers[tuple] = tuple_formatter
        yaml.Dumper.yaml_multi_representers[Enum] = enum_formatter

    setup_yaml_formatting()
    return yaml.dump(o, Dumper=CDumper)
