from django.conf.urls import url
from . import views

urlpatterns = [
    url(r'^$', views.release_build, name="release_build"),
]
