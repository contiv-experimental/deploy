FROM python:2.7
RUN mkdir -p /pycode
COPY . /pycode
WORKDIR /pycode
ENV https_proxy https://proxy.esl.cisco.com:8080
# RUN apt-get update
# RUN apt-get install -y tcpdump busybox fping iperf 
RUN pip install -r requirements.txt
CMD python app.py
