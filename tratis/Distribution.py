import warnings
import numpy as np
import scipy.stats as st
import ast
import json
import statistics


# Create models from data
def BestFitDistribution(data, isTime):
    # Convert time data to seconds from microseconds.

    if isTime == "true":
        data = [float(val) * 1e-6 for val in data]

    bins = int((len(data) * 1.0) / 10)

    if bins == 0:
        bins = len(data)
    """Model data by finding best fit distribution to data"""
    # Get histogram of original data
    y, x = np.histogram(data, bins=bins, density=True)
    x = (x + np.roll(x, -1))[:-1] / 2.0

    # Distributions to check
    DISTRIBUTIONS = [
        st.beta, st.chi2, st.expon, st.f, st.gamma, st.gumbel_r, st.invgamma,
        st.laplace, st.lognorm, st.norm, st.pareto, st.weibull_max
    ]

    # st.levy_stable,

    # Best holders
    best_distribution = st.norm
    best_params = (0.0, 1.0)
    best_sse = np.inf

    # Estimate distribution parameters from data
    for distribution in DISTRIBUTIONS:
        # Try to fit the distribution
        try:
            # Ignore warnings from data that can't be fit
            with warnings.catch_warnings():
                warnings.filterwarnings('ignore')

                # fit dist to data
                params = distribution.fit(data)

                # Separate parts of parameters
                arg = params[:-2]
                loc = params[-2]
                scale = params[-1]

                # Calculate fitted PDF and error with fit in distribution
                pdf = distribution.pdf(x, loc=loc, scale=scale, *arg)
                sse = np.sum(np.power(y - pdf, 2.0))

                # identify if this distribution is better
                if best_sse > sse > 0:
                    best_distribution = distribution
                    best_params = params
                    best_sse = sse

        except Exception:
            pass

    ret = {}
    ret['name'] = best_distribution.name
    ret['parameters'] = best_params
    ret['mean'] = np.mean(data)
    ret['sigma'] = statistics.stdev(data)
    return json.dumps(ret)