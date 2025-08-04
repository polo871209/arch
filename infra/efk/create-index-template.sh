#!/bin/bash

# Create Elasticsearch index template for kubernetes logs
# This ensures proper field mapping for parsed log fields

ELASTICSEARCH_HOST="https://elasticsearch-es-http.elastic-stack.svc.cluster.local:9200"
ELASTIC_PASSWORD=$(kubectl get secret elasticsearch-es-elastic-user -n elastic-stack -o jsonpath='{.data.elastic}' | base64 -d)

echo "Creating Elasticsearch index template for kubernetes logs..."

# Create index template with proper field mappings
curl -X PUT "${ELASTICSEARCH_HOST}/_index_template/kubernetes-logs" \
  -H "Content-Type: application/json" \
  -u "elastic:${ELASTIC_PASSWORD}" \
  -k \
  -d '{
    "index_patterns": ["kubernetes-*"],
    "priority": 500,
    "template": {
      "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
      },
      "mappings": {
        "properties": {
          "@timestamp": {
            "type": "date"
          },
          "level": {
            "type": "keyword"
          },
          "method": {
            "type": "keyword"
          },
          "path": {
            "type": "keyword"
          },
          "status_code": {
            "type": "integer"
          },
          "client_ip": {
            "type": "ip"
          },
          "client_port": {
            "type": "integer"
          },
          "protocol": {
            "type": "keyword"
          },
          "status_text": {
            "type": "text"
          },
          "msg": {
            "type": "text",
            "fields": {
              "keyword": {
                "type": "keyword"
              }
            }
          },
          "user_id": {
            "type": "keyword"
          },
          "email": {
            "type": "keyword"
          },
          "kubernetes": {
            "properties": {
              "pod_name": {
                "type": "keyword"
              },
              "namespace_name": {
                "type": "keyword"
              },
              "container_name": {
                "type": "keyword"
              },
              "pod_ip": {
                "type": "ip"
              }
            }
          },
          "cluster_name": {
            "type": "keyword"
          },
          "environment": {
            "type": "keyword"
          }
        }
      }
    }
  }'

echo ""
echo "✅ Index template created for kubernetes-* indices"