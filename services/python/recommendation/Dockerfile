FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . /app/recommendation/
ENV PYTHONPATH=/app
EXPOSE 8003
CMD ["uvicorn", "recommendation.main:app", "--host", "0.0.0.0", "--port", "8003"]
