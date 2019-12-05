from django.shortcuts import render


# Create your views here.
def alert(request):
    return render(request, "alert.html")
