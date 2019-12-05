from google.cloud import storage


def download_blob():
    """Downloads a blob from the bucket."""
    storage_client = storage.Client()
    bucket_name = 'istio-build/perf'
    bucket = storage_client.get_bucket(bucket_name)

    print('ID: {}'.format(bucket.id))
    print('Name: {}'.format(bucket.name))
    print('Storage Class: {}'.format(bucket.storage_class))
    print('Location: {}'.format(bucket.location))
    print('Location Type: {}'.format(bucket.location_type))
    print('Cors: {}'.format(bucket.cors))
    print('Default Event Based Hold: {}'
          .format(bucket.default_event_based_hold))
    print('Default KMS Key Name: {}'.format(bucket.default_kms_key_name))
    print('Metageneration: {}'.format(bucket.metageneration))
    print('Retention Effective Time: {}'
          .format(bucket.retention_policy_effective_time))
    print('Retention Period: {}'.format(bucket.retention_period))
    print('Retention Policy Locked: {}'.format(bucket.retention_policy_locked))
    print('Requester Pays: {}'.format(bucket.requester_pays))
    print('Self Link: {}'.format(bucket.self_link))
    print('Time Created: {}'.format(bucket.time_created))
    print('Versioning Enabled: {}'.format(bucket.versioning_enabled))
    print('Labels:')
    # bucket = storage_client.get_bucket(bucket_name)
    # blob = bucket.blob(source_blob_name)
    #
    # blob.download_to_filename(destination_file_name)
    #
    # print('Blob {} downloaded to {}.'.format(
    #     source_blob_name,
    #     destination_file_name))

download_blob()