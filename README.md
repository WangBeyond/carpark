# Instruction

## Init project
``` bash
kubectl apply -f deploy
kubectl port-forward $(kubectl get po -l app=mysql -o Name) 7000:3306
mysql -h 127.0.0.1 -uroot -ppassword -P 7000 < ./data/db-init.sql
```

## To update availability
eg: 
``` bash
curl http://192.168.99.100:30000/carparks/update
```

## To query nearest available carparks
eg: 
``` bash
curl http://192.168.99.100:30000/carparks/nearest?latitude=1.37326&longitude=103.897&page=4&per_page=5
```
