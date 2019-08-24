from __future__ import print_function

from google.cloud import dns
import os
import sys
import argparse
import ipaddress


class DNSException(Exception):
    pass


def find_rs(rs, record_type, name=None):
    for r in rs:
        if r.record_type != record_type:
            continue

        if name is None or r.name == name:
            return r

    return None


def remove_rec(changes, rec):
    if rec is not None:
        changes.delete_record_set(rec)

# creates
# A record to ingress-${domain}. --> IP
# CNAME *.domain --> ingess-${domain}
# PTR record from IP.in-addr.arpa. --> ingress-${domain}


def register_addr(zone, ip, domain):

    if not domain.endswith("."):
        domain += "."

    rs = list(zone.list_resource_record_sets())

    # get the domain managed by this zone
    soa_rec = find_rs(rs, "SOA")
    if soa_rec is None:
        raise DNSException("Could not get SOA record for zone " + zone.name)

    zone_domain = soa_rec.name

    if not domain.endswith(zone_domain):
        raise DNSException(domain + " must be a subdomain of " + zone_domain)

    ingress = "ingress-" + domain
    cname = "*." + domain
    ptr = ".".join(reversed(ip.split('.'))) + ".{}".format(zone_domain)

    ingress_rec = find_rs(rs, "A", ingress)
    cname_rec = find_rs(rs, "CNAME", cname)
    ptr_rec = find_rs(rs, "TXT", ptr)

    changes = zone.changes()

    remove_rec(changes, ingress_rec)
    remove_rec(changes, cname_rec)
    remove_rec(changes, ptr_rec)

    changes.add_record_set(zone.resource_record_set(ingress, 'A', 300, [ip]))
    changes.add_record_set(zone.resource_record_set(
        cname, 'CNAME', 300, [ingress]))
    changes.add_record_set(
        zone.resource_record_set(ptr, 'TXT', 300, [ingress]))

    str_changes = changes._build_resource()
    try:
        changes.create()
    except Exception as ex:
        print(ex)
        print(str_changes)


def get_zone(zone_name, project_id=None):
    project_id = project_id or os.environ['PROJECT_ID']

    client = dns.Client(project=project_id)
    zone = client.zone(zone_name)
    if not zone.exists():
        raise DNSException(zone_name + " does not exist")

    return zone


def main(args):
    if args.project_id is None:
        args.project_id = os.environ['PROJECT_ID']

    zone = get_zone(args.zone_name, args.project_id)

    try:
        register_addr(zone, args.ingress_ip, args.subdomain)
    except DNSException as ex:
        print(ex)
        return -1

    return 0


def validate_ip(addr):
    # throws exception for invalid ip addresses
    ip = ipaddress.ip_address(unicode(addr))

    return addr


def get_parser():
    parser = argparse.ArgumentParser(
        description="Register dns in a standard way")

    parser.add_argument(
        "zone_name", help="Cloud dns zone name, example:qual-stio-org")
    parser.add_argument("subdomain", help="subdomain to associate with the ip")
    parser.add_argument("ingress_ip", type=validate_ip)
    parser.add_argument("--project-id")

    return parser


if __name__ == "__main__":
    parser = get_parser()
    args = parser.parse_args()
    sys.exit(main(args))

# example:
# python ./update_dns.py qual-stio-org  cls1302.v13.qualistio.org
# 35.202.78.224
