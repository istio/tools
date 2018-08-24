import hashlib


def hex(path: str) -> str:
    """Returns the hex of the MD5 hash of the file at path's contents."""
    hash_md5 = hashlib.md5()
    with open(path, 'rb') as f:
        for chunk in iter(lambda: f.read(4096), b''):
            hash_md5.update(chunk)
    return hash_md5.hexdigest()
