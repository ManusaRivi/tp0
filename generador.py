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
        'environment': ['PYTHONUNBUFFERED=1'],
        'networks': ['testing_net'],
        'volumes': ['./server/config.ini:/config.ini']
    }
    for i in range(1, cantidad_clientes + 1):
        services[f'client{i}'] = {
            'container_name': f'client{i}',
            'image': 'client:latest',
            'entrypoint': '/client',
            'environment': [f'CLI_ID={i}'],
            'networks': ['testing_net'],
            'volumes': ['./client/config.yaml:/config.yaml', f'./.data/agency-{i}.csv:/agency.csv'],
            'depends_on': ['server'],
            'env_file': [f'./agencies/{i}.env']
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