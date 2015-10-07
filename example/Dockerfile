FROM python:2.7
RUN mkdir -p /pycode
COPY . /pycode
WORKDIR /pycode
RUN apt-get update
RUN apt-get install -y tcpdump fping iperf 
RUN pip install -r requirements.txt
CMD python app.py
