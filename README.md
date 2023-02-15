# hand_made_autoscaling_system
> It is a proof of concept for [this paper](https://docs.google.com/document/d/1D409FtuKj8ayKGgy4r7Io-QS9bW0WdU8/edit?usp=sharing&ouid=106853607019198247243&rtpof=true&sd=true)

## Description
It interacts with Firestore to auto-scale servers based on the values stored in Firestore. <br>
User sends request to run his/her task. Task only represents the size at the moment. <br>
Server has a capacity for the total size of tasks it can maintain. Therefore, it auto-scales if 
there are more tasks that one server cannot handle. 

## How to run
1. It is important to note that user of this program must create her/his own project and run below command to register access token
```shell
gcloud auth application-default login
```

2. Git clone this repository
3. Create a foundation of the Firestore column.
   1. Add `threshold`, `tasks`, `servers` in Collection 
   2. In the `threshold` column, document should be also `threshold`, and data should have `soft max`:20 and `max`:30
   3. Values of `soft max` and `max` can be changed based on the program user's preference.