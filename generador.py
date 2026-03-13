import sys

# For now, returns the same compose.
# TODO: generate a compose file with the specified amount of clients.
def generate_compose(compose_base, cantidad_clientes):
    compose_content = compose_base
    return compose_content

def main():
    if len(sys.argv) != 3:
        print("Uso: python3 generar_compose.py <compose_base> <client_amount>")
        sys.exit(1)

    base_compose = sys.argv[1]
    client_amount = int(sys.argv[2])

    compose_content = generate_compose(base_compose, client_amount)
    print(compose_content)

main()