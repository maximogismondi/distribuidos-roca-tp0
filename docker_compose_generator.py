import sys

NAME = "tp0"


def indent(text, level):
    return "  " * level + text


class Service:
    def __init__(
        self,
        container_name,
        image,
        entrypoint,
        environment,
        networks,
        depends_on,
        volumes,
    ):
        self.container_name = container_name
        self.image = image
        self.entrypoint = entrypoint
        self.environment = environment
        self.networks = networks
        self.depends_on = depends_on
        self.volumes = volumes

    def __str__(self):
        lines = []
        lines.append(indent(f"{self.container_name}:", 1))
        lines.append(indent(f"container_name: {self.container_name}", 2))
        lines.append(indent(f"image: {self.image}", 2))
        lines.append(indent(f"entrypoint: {self.entrypoint}", 2))

        if self.environment:
            lines.append(indent("environment:", 2))
            for key, value in self.environment.items():
                lines.append(indent(f"- {key}={value}", 3))

        if self.networks:
            lines.append(indent("networks:", 2))
            for network in self.networks:
                lines.append(indent(f"- {network}", 3))

        if self.depends_on:
            lines.append(indent("depends_on:", 2))
            for dep in self.depends_on:
                lines.append(indent(f"- {dep}", 3))

        if self.volumes:
            lines.append(indent("volumes:", 2))
            for key, value in self.volumes.items():
                lines.append(indent(f"- {key}:{value}", 3))

        return "\n".join(lines) + "\n"


class Network:
    def __init__(self, network_name, driver, subnets):
        self.network_name = network_name
        self.driver = driver
        self.subnets = subnets

    def __str__(self):
        lines = []
        lines.append(indent(f"{self.network_name}:", 1))
        lines.append(indent("ipam:", 2))
        lines.append(indent(f"driver: {self.driver}", 3))
        lines.append(indent("config:", 3))
        for subnet in self.subnets:
            lines.append(indent(f"- subnet: {subnet}", 4))

        return "\n".join(lines) + "\n"


def get_args():
    try:
        output_file = sys.argv[1]
        n_clients = int(sys.argv[2])
    except IndexError:
        print("Usage: python docker-compose-generator.py <output_file> <n_clients>")
        sys.exit(1)
    except ValueError:
        print("Number of clients must be an integer")
        sys.exit(1)

    return output_file, n_clients


def generate_docker_compose(output_file, n_clients):

    server = Service(
        container_name="server",
        image="server:latest",
        entrypoint="python3 /main.py",
        environment={"PYTHONUNBUFFERED": "1", "LOGGING_LEVEL": "INFO"},
        networks=["testing_net"],
        depends_on=[],
        volumes={"./server/config.ini": "/config.ini"},
    )

    client_services = [
        Service(
            container_name=f"client{i}",
            image="client:latest",
            entrypoint="/client",
            environment={"CLI_ID": f"{i}", "CLI_LOG_LEVEL": "INFO"},
            networks=["testing_net"],
            depends_on=["server"],
            volumes={"./client/config.yaml": "/config.yaml"},
        )
        for i in range(1, n_clients + 1)
    ]

    network = Network(
        network_name="testing_net", driver="default", subnets=["172.25.125.0/24"]
    )

    with open(output_file, "w") as f:
        f.write(f"name: {NAME}\n")
        f.write("services:\n")
        f.write(str(server) + "\n")
        for client in client_services:
            f.write(str(client) + "\n")
        f.write("networks:\n")
        f.write(str(network))


def main():
    output_file, n_clients = get_args()
    generate_docker_compose(output_file, n_clients)


if __name__ == "__main__":
    main()
