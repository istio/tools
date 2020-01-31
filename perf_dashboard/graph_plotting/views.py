from django.shortcuts import render
from django.core.files.storage import FileSystemStorage


def graph_plotting(request):
    if request.method == 'POST' and request.FILES.get('myfile'):
        myfile = request.FILES.get('myfile')
        fs = FileSystemStorage()
        filename = fs.save(myfile.name, myfile)
        uploaded_file_url = fs.url(filename)
        return render(request, 'graph_plotting.html', {
            'uploaded_file_url': uploaded_file_url
        })
    return render(request, 'graph_plotting.html')
