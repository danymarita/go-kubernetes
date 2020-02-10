This apps based on article https://www.callicoder.com/deploy-containerized-go-app-kubernetes/

How to run

1. Run minikube ssh to create hostpath /tmp/mysqldata defined in mysql-volume.yml
2. Run Mysql kubernetes apps in folder mysql
   - mysql-storage-class.yml
   - mysql-volume.yml
   - mysql-volume-claim.yml
   - mysql-secrets.yml
   - mysql-deployment.yml
   - mysql-service.yml
3. Enter mysql pods then create database -> create database go_kubernetes;
4. Create table
CREATE TABLE `products` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `code` varchar(255) DEFAULT NULL,
  `price` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_products_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=latin1;

5. Run kubectl apply -f go-kubernetes-apps-configmap.yml
6. Run kubectl apply -f go-kubernetes-apps-deployment.yml
7. Check deployment, pods and service. Make sure all running well
8. Run kubectl cluster-info to get IP cluster
9. Test hit apps using http://cluster-ip:service-port/
10. List Endpoint
   - /
   - /healt-check
   - /readiness
   - /products --> list products
   - /product/create --> create product

Note:
Change namespace on cluster local kubectl config set-context --current --namespace=go-kubernetes-apps
If apps change, create image with new tag then upload to docker hub and change go-kubernetes-apps-deployment.yml container image
1. Build image with new tag
docker build -t danympradana/go-kubernetes:1.1.0 .
2. Login to docker hub
docker login
3. Push image
docker push danympradana/go-kubernetes:1.1.0