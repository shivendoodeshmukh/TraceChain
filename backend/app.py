from flask import Flask, render_template, request
import logging
import os

app = Flask(__name__)

@app.route('/')
def index():
    return 'Invalid Request'
    
@app.route('/api/enroll', methods=['POST'])
def enroll():
    logging.info('Enroll')
    # Enroll Device onto the Blockchain Network
    
    return 'Enroll'


if __name__ == '__main__':
    app.run()