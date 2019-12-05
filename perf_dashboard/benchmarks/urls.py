from django.conf.urls import url
from . import views

urlpatterns = [
    url('cpu_memory/', views.cpu_memory, name="cpu_memory"),
    url('latency/', views.latency, name="latency"),
    url('flame_graph/', views.flame_graph, name="flame_graph"),
    url('micro_benchmarks/', views.micro_benchmarks, name="micro_benchmarks"),
]