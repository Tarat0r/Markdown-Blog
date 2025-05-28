#!/bin/bash 
# kubectl -n monitoring port-forward svc/mon-grafana 3000:80

kubectl -n mdblog run curl --rm -i --restart=Never --image=curlimages/curl -- \
  sh -c "while true; do curl -i -H 'Authorization: xFRFXi9fn3xwsOS8di8abUfouKJs6HqrrdejCkW1cvTm9XiCo6B49YeZL9HvHvdK' http://mdblog-api:80/notes; done"
