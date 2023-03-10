FROM python:3

WORKDIR app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY server.py .
ENTRYPOINT ["python3", "server.py"]
