import requests
import time
import socketserver
from http.server import BaseHTTPRequestHandler
import threading

num_requests = 50
requests_received = 0
start_time = time.time()
class MyHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        global requests_received
        requests_received += 1
        print("Received reply:", requests_received)

        if requests_received == num_requests:
            total_time = time.time() - start_time
            print("Time:", total_time, "seconds")

        self.send_response(200)

def startserver():
    httpd = socketserver.TCPServer(("", 8080), MyHandler)
    httpd.serve_forever()
  


url = "http://localhost:9001/app/manufacturer"

x = threading.Thread(target=startserver, daemon=True)
x.start()
print("Start server")
time.sleep(10)
for i in range(num_requests):
    payload = "{\n\t\"to_application\":\"SUPPLIER\",\n\t\"num_units_to_sell\":20,\n\t\"amount_to_be_collected\":100,\n\t\"transaction_type\": \"GLOBAL_TXN\"\n}"
    headers = {
        'Content-Type': "application/json",
        'cache-control': "no-cache",
        }

    response = requests.request("POST", url, data=payload, headers=headers)

x.join()
