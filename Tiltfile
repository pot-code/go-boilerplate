k8s_yaml(['k8s-pod.yaml','k8s-svc.yaml'])

# Build: tell Tilt what images to build from which directories
docker_build('potdockercode/go-boilerplate-backend', '.')

k8s_resource('mysql', port_forwards='3306')
k8s_resource('redis', port_forwards='6379')
k8s_resource('go-boilerplate-backend', port_forwards='8081')