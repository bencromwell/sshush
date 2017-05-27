import yaml


def read_file(file):
    with open(file, 'r') as stream:
        # try:
            return yaml.load(stream)
        # except yaml.YAMLError as e:
        #     print(e)


def process_yaml(ssh_config_yaml):
    defaults = extract_section('default', ssh_config_yaml, {})
    global_values = extract_section('global', ssh_config_yaml)

    for identifier, item in ssh_config_yaml.items():
        write('# {}'.format(identifier))
        hosts = item['Hosts']
        del item['Hosts']

        settings = {**defaults, **item}

        for reference, host_details in hosts.items():
            write(reference)
            host_settings = {**settings, **host_details}

            for k, v in host_settings.items():
                write('    {} {}'.format(k, v))

            write("\n")

    if global_values is not None:
        write('Host *')
        for key, value in global_values.items():
            write('    {} {}'.format(key, value))


def extract_section(section_name, ssh_config, default_to = None):
    if ssh_config[section_name]:
        default_to = ssh_config[section_name]
        del ssh_config[section_name]
    return default_to


def write(text):
    print(text)
