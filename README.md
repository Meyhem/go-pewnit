# go-pewnit
*Simple Web app stress testing tool which can utilize Http flood, Connection flood and Slowloris attacks.*

Control strength of attack via -c, --concurrency option

# Usage

Run ./test_target.py for localhost:8080 test server.

**Simple HTTP flood with 200 paralel attacks:**

./pewnit http://localhost:8080 -c 200 -a httpflood

**Connection flood**

./pewnit http://localhost:8080 -c 200 -a connectionflood

**Slowloris**

./pewnit http://localhost:8080 -c 200 -a slowloris

**POST method**

./pewnit http://localhost:8080 -c 200 -a slowloris -m POST

**POST with data**

Note that Content-Type mime header must be defined by user, otherwise its considered invalid HTTP and your attack might not have effect.

./pewnit http://localhost:8080 -c 200 -a httpflood -b username="default&password=default" --header Content-Type:application/x-www-form-urlencoded

*Common content types:* 
* application/x-www-form-urlencoded  
* application/json  
* application/javascript   
* application/octet-stream   
* application/xml   
* multipart/form-data  
* text/css    
* text/csv    
* text/html    
* text/plain    
* text/xml  
* ...

For more see **./pewnit -h**

For education purposes only. 
