#!/usr/bin/env python
import cv2
import imutils
import time
import sys
import argparse
import base64
import simplejson as json

def videoStreamer(path, width=800, height=None, skip=None, js=False):
    stream = cv2.VideoCapture(path)
    frames = int(stream.get(cv2.CAP_PROP_FRAME_COUNT))        
    FPS = stream.get(cv2.CAP_PROP_FPS)
        
    if skip == None:
        skip = int(FPS)
    frame = None
    while True:        
        for i in range(skip):
            stream.grab()
        (grabbed, frame) = stream.read()
        if not grabbed:
            return                
        frame = imutils.resize(frame, width=width, height=height)
        f = stream.get(cv2.CAP_PROP_POS_FRAMES)
        t = stream.get(cv2.CAP_PROP_POS_MSEC)
        res = bytearray(cv2.imencode(".jpeg", frame)[1])
        size = str(len(res))

        if js:            
            obj = {"frame": int(f), "millis": int(t), "total": frames, "image": base64.b64encode(res)}
            sys.stdout.write(json.dumps(obj))
            sys.stdout.write("\n")
        else:        
            sys.stdout.write("Content-Type: image/jpeg\r\n")
            sys.stdout.write("Content-Length: " + size + "\r\n\r\n")
            sys.stdout.write(res)
            sys.stdout.write("\r\n")
            sys.stdout.write("--informs\r\n")

if __name__ == '__main__':
    parser = argparse.ArgumentParser()    
    parser.add_argument("--path", help="path of the video",type=str, required=True)
    parser.add_argument("--skip", help="number of frames to skip", default=None, type=int)
    parser.add_argument("--width", help="width to redimension the stream", default=600, type=int)
    parser.add_argument("--height", help="height to redimension the stream", default=None, type=int)
    parser.add_argument("--json", help="return json instead of piping the stream", type=bool, default=False)
    args = parser.parse_args()        
    videoStreamer(args.path, width=args.width, height=args.height, skip=args.skip, js=args.json)
    
