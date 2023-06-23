# 1 -> primeira versão com volume de app externo
tag=1

docker build . --tag=registry.lab.systemhaus.com.br/prometheus-postgres-exporter:${tag}
docker push registry.lab.systemhaus.com.br/prometheus-postgres-exporter:${tag}
