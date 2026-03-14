import sys
import yaml

def generate_compose(cantidad_clientes):
    compose_content = {}
    compose_content['name'] = 'tp0'
    
    services = {}

    services['server'] = {
        'container_name': 'server',
        'image': 'server:latest',
        'entrypoint': 'python3 /main.py',
        'environment': ['PYTHONUNBUFFERED=1', 'LOGGING_LEVEL=DEBUG'],
        'networks': ['testing_net']
    }
    for i in range(1, cantidad_clientes + 1):
        services[f'client{i}'] = {
            'container_name': f'client{i}',
            'image': 'client:latest',
            'entrypoint': '/client',
            'environment': [f'CLI_ID={i}', 'CLI_LOG_LEVEL=DEBUG'],
            'networks': ['testing_net'],
            'depends_on': ['server']
        }
    compose_content['services'] = services

    compose_content['networks'] = {
        'testing_net': {
            'ipam': {
                'driver': 'default',
                'config': [
                    {'subnet': '172.25.125.0/24'}
                ]
            }
        }
    }
    yaml.dump(compose_content, sys.stdout)

def main():
    if len(sys.argv) != 2:
        sys.exit(1)

    client_amount = int(sys.argv[1])

    generate_compose(client_amount)

main()