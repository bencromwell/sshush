import argparse
from os.path import expanduser
from sshush.sshush import read_file, process_yaml


def main():
    default_path = '{home}/.ssh/config'.format(home=expanduser('~'))
    default_yaml_path = '{home}/.ssh/config.yml'.format(home=expanduser('~'))

    arg_parser = argparse.ArgumentParser()

    arg_parser.add_argument(
        '--source', '-s',
        help='Path to source YAML file if it differs from {}'.format(default_yaml_path),
        default=default_yaml_path
    )

    arg_parser.add_argument(
        '--path', '-p',
        help='Path to SSH config file if it differs from {}'.format(default_path),
        default=default_path
    )

    args = arg_parser.parse_args()

    print('sshush running with path "{path}" and source YAML "{yaml}"'.format(
        path=args.path,
        yaml=args.source
    ))

    yaml_obj = read_file(args.source)
    config_file_contents = process_yaml(yaml_obj)

    try:
        with open(args.path, 'w') as fh:
            fh.write(config_file_contents)
            fh.write("\n")
            print('{} written successfully'.format(args.path))
    except IOError as exc:
        print('Error:', exc.strerror)
        exit(1)


if __name__ == '__main__':
    main()
