# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from django.shortcuts import render
from django.core.files.storage import FileSystemStorage


def graph_plotting(request):
    if request.method == 'POST' and request.POST.get('my_benchmark_type') and request.POST.get('my_graph_name') and \
            request.FILES.get('my_file'):
        my_benchmark_type = request.POST.get('my_benchmark_type')
        my_graph_name = request.POST.get('my_graph_name')
        my_file = request.FILES.get('my_file')
        fs = FileSystemStorage()
        filename = fs.save(my_file.name, my_file)
        uploaded_file_url = fs.url(filename)
        context = {
            'uploaded_file_url': uploaded_file_url,
            'user_benchmark_type': my_benchmark_type,
            'user_graph_name': my_graph_name,
        }
        return render(request, 'graph_plotting.html', context=context)
    return render(request, 'graph_plotting.html')
