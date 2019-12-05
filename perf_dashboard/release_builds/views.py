from django.shortcuts import render


# Create your views here.
def release_build(request):
    return render(request, "release_builds.html")
