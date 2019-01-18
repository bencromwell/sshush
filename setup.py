from setuptools import setup

setup(name='sshush',
      version='1.3.3',
      description='SSH Config management tool',
      url='http://github.com/bencromwell/sshush',
      author='Ben Cromwell',
      author_email='placeholder@example.com',
      license='MIT',
      packages=['sshush'],
      install_requires=[
          'pyaml',
      ],
      zip_safe=False,
      test_suite='nose.collector',
      tests_require=['nose'],
      entry_points={
          'console_scripts': [
              'sshush = sshush.command_line:main'
          ],
      }
)
