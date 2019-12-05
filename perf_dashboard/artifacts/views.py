from django.shortcuts import render


# Create your views here.
def artifact(request):
    return render(request, "artifact.html")
