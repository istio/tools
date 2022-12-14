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

import os
from django.shortcuts import render

#current_release = [os.getenv('CUR_RELEASE')]


# Create your views here.
def settings(request):

    current_release = request.COOKIES.get('currentRelease')
    project_id = request.COOKIES.get('projectId')
    bucket_name = request.COOKIES.get('bucketName')
    download_dataset_days = request.COOKIES.get('downloadDatasetDays')

    if not current_release:
        current_release = os.getenv('CUR_RELEASE')
    if not project_id:
        project_id = os.getenv('PROJECT_ID')
    if not bucket_name:
        bucket_name = os.getenv('BUCKET_NAME')
    if not download_dataset_days:
        download_dataset_days = os.getenv('DOWNLOAD_DATASET_DAYS')

    print(current_release, project_id, bucket_name, download_dataset_days)

    return render(request, 'settings.html', {'current_release': current_release,
                                             'project_id': project_id,
                                             'bucket_name': bucket_name,
                                             'download_dataset_days': download_dataset_days})
