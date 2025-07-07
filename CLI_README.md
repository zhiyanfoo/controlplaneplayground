# Resource Manager CLI

This CLI tool allows you to manage Envoy resources through the ResourceManager gRPC service by reading configuration from YAML files.

## Building the CLI

```bash
# From the project root
./scripts/build_cli.sh
```

Or manually:
```bash
go build -o cli/cli cli/main.go
```

## Usage

```bash
./cli/cli -config <yaml_file> -action <update|delete> [-server <address>]
```

### Flags

- `-config`: Path to YAML configuration file (required)
- `-action`: Action to perform: `update` or `delete` (default: `update`)
- `-server`: gRPC server address (default: `localhost:18000`)

## YAML Configuration Format

The YAML file should have the following structure:

```yaml
server_address: "localhost:18000"  # Optional, can be overridden by -server flag
resources:
  - type_url: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
    name: "xdstp:///envoy.config.cluster.v3.Cluster/my_cluster"
    data: |
      {
        "name": "xdstp:///envoy.config.cluster.v3.Cluster/my_cluster",
        "connect_timeout": "5s",
        "type": "EDS",
        "eds_cluster_config": {
          "eds_config": {
            "ads": {}
          }
        },
        "lb_policy": "ROUND_ROBIN"
      }
  
  - type_url: "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment"
    name: "xdstp:///envoy.config.cluster.v3.Cluster/my_cluster"
    data: |
      {
        "cluster_name": "xdstp:///envoy.config.cluster.v3.Cluster/my_cluster",
        "endpoints": [
          {
            "lb_endpoints": [
              {
                "endpoint": {
                  "address": {
                    "socket_address": {
                      "address": "127.0.0.1",
                      "port_value": 50051
                    }
                  }
                }
              }
            ]
          }
        ]
      }
```

### Fields

- `type_url`: The protobuf type URL for the resource (e.g., `type.googleapis.com/envoy.config.cluster.v3.Cluster`)
- `name`: The name of the resource (should match the xDSTP format used in the control plane)
- `data`: The serialized resource data (JSON format for Envoy resources)

## Examples

### Update Resources

```bash
./cli/cli -config examples/update_resources.yaml -action update
```

### Delete Resources

```bash
./cli/cli -config examples/delete_resources.yaml -action delete
```

### Custom Server Address

```bash
./cli/cli -config examples/update_resources.yaml -action update -server localhost:18001
```

## Example YAML Files

- `examples/update_resources.yaml`: Example configuration for updating clusters and endpoints
- `examples/delete_resources.yaml`: Example configuration for deleting resources

## Prerequisites

1. The control plane must be running with the ResourceManager service enabled
2. The control plane should be listening on the specified gRPC port (default: 18000)

## Notes

- The `data` field for delete operations is ignored but should be present in the YAML structure
- Resource names should follow the xDSTP format used by the control plane
- The CLI will process all resources in the YAML file sequentially
- Each resource operation is logged with success/failure status 
