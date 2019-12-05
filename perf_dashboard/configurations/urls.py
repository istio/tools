from django.conf.urls import url
from . import views

urlpatterns = [
    url(r'^$', views.configuration, name="configuration"),
]