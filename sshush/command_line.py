import argparse
from os.path import expanduser
from sshush.parser import read_file, process_yaml


def main():
    default_path = '{home}/.ssh/test-config'.format(home=expanduser('~'))

    arg_parser = argparse.ArgumentParser()

    arg_parser.add_argument('yaml_file', help='Source YAML')

    arg_parser.add_argument(
        '--path', '-p',
        help='Path to SSH config file if it differs from {}'.format(default_path),
        default=default_path
    )

    args = arg_parser.parse_args()

    print('sshush running with path "{path}" and source YAML "{yaml}"'.format(
        path=args.path,
        yaml=args.yaml_file
    ))

    yaml_obj = read_file(args.yaml_file)
    config_file_contents = process_yaml(yaml_obj)

    with open(args.path, 'w') as fh:
        try:
            fh.write(config_file_contents)
            fh.write("\n")
            print('{} written successfully'.format(args.path))
        except IOError as exc:
            print(exc.strerror)
            exit(1)


if __name__ == '__main__':
    main()
