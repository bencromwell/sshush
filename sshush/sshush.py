import yaml
from collections import OrderedDict


# https://stackoverflow.com/a/21912744/149998
def ordered_load(stream, loader=yaml.Loader, object_pairs_hook=OrderedDict):
    class OrderedLoader(loader):
        pass

    def construct_mapping(loader, node):
        loader.flatten_mapping(node)
        return object_pairs_hook(loader.construct_pairs(node))

    OrderedLoader.add_constructor(
        yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
        construct_mapping
    )

    return yaml.load(stream, OrderedLoader)


def read_file(file):
    with open(file, 'r') as stream:
        # try:
            return ordered_load(stream, yaml.SafeLoader)
        # except yaml.YAMLError as e:
        #     print(e)


def process_yaml(ssh_config_yaml):
    output = []

    defaults = extract_section('default', ssh_config_yaml, {})
    global_values = extract_section('global', ssh_config_yaml)

    configs = {}

    for identifier, item in ssh_config_yaml.items():
        output.append('# {}'.format(identifier))
        hosts = item['Hosts']

        if 'Config' not in item:
            item['Config'] = {}

        if 'Extends' in item and item['Extends'] in configs:
            settings = {**defaults, **configs[item['Extends']], **item['Config']}
        else:
            settings = {**defaults, **item['Config']}

        configs[identifier] = settings

        # ugly remapping to handle a list of hosts
        # as per the ciscos.yml example
        # ciscos:
        #   Config:
        #     Ciphers: aes128-ctr,aes192-ctr,aes256-ctr,aes128-cbc,3des-cbc
        #     KexAlgorithms: +diffie-hellman-group1-sha1
        #     HostKeyAlgorithms: ssh-rsa,ssh-dss
        #     PubkeyAuthentication: "no"
        #   Hosts:
        #     - fooas*.adm
        #     - foocs*.adm
        #     - foocr01.adm
        #     - cs*.foo.adm
        #
        if isinstance(hosts, (list,)):
            tmp_hosts = hosts
            hosts = {}
            for host in tmp_hosts:
                hosts[host] = host

        for reference, host_details in hosts.items():
            output.append('Host {}'.format(reference))

            # if it's a string then it's a straight reference to IP or hostname mapping
            if isinstance(host_details, str):
                host_details = {
                    'HostName': host_details
                }

            host_settings = {**settings, **host_details}

            for k, v in host_settings.items():
                output.append('    {} {}'.format(k, v))

            output.append("")

    if global_values is not None:
        output.append('Host *')
        for key, value in global_values.items():
            output.append('    {} {}'.format(key, value))

    return "\n".join(output)


def extract_section(section_name, ssh_config, default_to = None):
    if section_name in ssh_config:
        default_to = ssh_config[section_name]
        del ssh_config[section_name]

    return default_to
