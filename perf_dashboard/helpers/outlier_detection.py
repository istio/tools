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


def find_outliers(pattern_data):
    perf_val = []
    """
    Example of an array in pattern_data: ['2019', '12', '23', 4.478, '20191223', 'release-1.4.20191223-16.da6d4af0a2e1c3207edfd97f09d07a638c59e89a']
    """
    for data in pattern_data:
        perf_num = data[3]
        if perf_num != 'null':
            release_date = data[4]
            release_name = data[5]
            perf_val.append((perf_num, release_date, release_name))

    perf_val.sort(key=lambda x: x[0])   # sort by perf_number
    lower_inner_fence, upper_inner_fence, lower_outer_fence, upper_outer_fence = calculate_fences(perf_val)
    mild_outliers = []
    extreme_outliers = []
    outliers = []
    for val in perf_val:
        if ((val[0] > upper_inner_fence and val[0] < upper_outer_fence) or (val[0] < lower_inner_fence and val[0] > lower_outer_fence)):
            mild_outliers.append((val[0], val[1], val[2]))
        elif val[0] > upper_outer_fence or val[0] < lower_outer_fence:
            extreme_outliers.append((val[0], val[1], val[2]))
    outliers = mild_outliers + extreme_outliers
    return outliers


def calculate_fences(perf_val):
    Q1, Q3, IQR = cacluate_interquartile_range(perf_val)
    lower_inner_fence = Q1 - 1.5 * IQR
    upper_inner_fence = Q3 + 1.5 * IQR
    lower_outer_fence = Q1 - 3 * IQR
    upper_outer_fence = Q3 + 3 * IQR
    return lower_inner_fence, upper_inner_fence, lower_outer_fence, upper_outer_fence


def cacluate_interquartile_range(perf_val):
    # First Quartile
    Q1 = 0
    # Third Quartile
    Q3 = 0
    # Intequartile Range
    IQR = Q3 - Q1

    length = len(perf_val)
    if length % 2 == 0:
        Q1 = perf_val[int(length / 4)][0]
        Q3 = perf_val[int(length / 2) + int(length / 4)][0]
        IQR = Q3 - Q1
    else:
        Q1 = perf_val[int(length / 4)][0]
        Q3 = perf_val[int(length / 2) + int(length / 4) + 1][0]
        IQR = Q3 - Q1

    return Q1, Q3, IQR


def format_outliers(outliers, mixer_mode):
    """
    outliers is a tuple of array, a tuple example:
    (2.634, '20200105', 'release-1.4.20200105-16.c6691c61f845791e58ab42ede6cd836da93aa63a')
    """
    formatted_outlier = [[]] * len(outliers)
    if len(outliers) > 0:
        istio_git_commit_url = "https://github.com/istio/istio/commit/"
        for i in range(len(outliers)):
            formatted_outlier[i] = [0] * 4
            perf_num = outliers[i][1]
            formatted_outlier[i][0] = perf_num
            release_name = outliers[i][2]
            commit_sha = outliers[i][2].split('-')[1].split('.')[1]
            if outliers[i][2].startswith('master'):
                formatted_outlier[i][1] = istio_git_commit_url + commit_sha
            else:
                formatted_outlier[i][1] = istio_git_commit_url + commit_sha
            formatted_outlier[i][2] = release_name
            formatted_outlier[i][3] = (mixer_mode, perf_num)
    return formatted_outlier
