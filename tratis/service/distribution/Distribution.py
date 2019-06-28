import logging
import time
from jaeger_client import Config

if __name__ == "__main__":
    log_level = logging.DEBUG
    logging.getLogger('').handlers = []
    logging.basicConfig(format='%(asctime)s %(message)s', level=log_level)

    config = Config(
        config={ # usually read from some yaml config
            'sampler': {
                'type': 'const',
                'param': 100,
            },
            'local_agent': {
                'reporting_host': 'localhost',
                'reporting_port': '15032',
            },
            'logging': True,
        },  
        service_name='bookinfo',
        validate=True,
    )
    # this call also sets opentracing.tracer
    tracer = config.initialize_tracer()

    print tracer.__dict__

    # with tracer.start_span('TestSpan') as span:
    #     span.log_kv({'event': 'test message', 'life': 42})

    #     with tracer.start_span('ChildSpan', child_of=span) as child_span:
    #         span.log_kv({'event': 'down below'})

    time.sleep(2)   # yield to IOLoop to flush the spans - https://github.com/jaegertracing/jaeger-client-python/issues/50
    tracer.close()  # flush any buffered spans