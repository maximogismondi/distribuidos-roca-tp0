import sys
import docker_compose_generator as dcg

TEST_TEXT = "hello world"


def generate_docker_compose():
    server = dcg.Service(
        container_name="server",
        image="server:latest",
        entrypoint="python3 /main.py",
        environment={"PYTHONUNBUFFERED": "1", "LOGGING_LEVEL": "DEBUG"},
        networks=["testing_net"],
        depends_on=[],
        volumes={"./server/config.ini": "/config.ini"},
    )

    TEST_CMD = f'sh -c "echo \\"{TEST_TEXT}\\" | nc server 12345 | grep -q \\"^{TEST_TEXT}$\\""'

    client = dcg.Service(
        container_name="test_client",
        image="alpine:latest",
        entrypoint=TEST_CMD,
        networks=["testing_net"],
        depends_on=["server"],
        environment={},
        volumes={},
    )

    network = dcg.Network(
        network_name="testing_net", driver="default", subnets=["172.25.125.0/24"]
    )

    services = [server, client]

    return services, network


def get_args():
    try:
        output_file = sys.argv[1]
        return output_file
    except IndexError:
        print("Usage: python validate-echo-server.py <output_file>")
        sys.exit(1)


def main():
    output_file = get_args()
    services, networks = generate_docker_compose()
    dcg.write_to_file(output_file, services, networks)


if __name__ == "__main__":
    main()
