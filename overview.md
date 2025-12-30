I am building a gym application meant for noobs. It will first be a simple logger for exercises. It is not meant tio replace Fitbit or similar tracking apps. however, it is meant to help you know where you are at with weights and to know where you are going. 

The app will work as follows: 
- auth: username and password. No email. 
- add exercise: name, muscle group, equipment, description, instructions, video link, and image link. 
- Log exercise: name, sets, reps, weight, and date. 

The app will then show your last weight for a specific exercise and allow you to log a new weight. The app will also show a line chart showing your weight progress for a specific exercise. 

The app will be vue3 with shadcn-vue for the UI. 

The app will use an express backend with a sqlite database.

The app will take into account: 
- User should only be able to see and edit their own exercises. 
- User should be able to see and edit their own workout logs. 
- User should be able to see and edit their own weight progress. 
- User should be able to see and edit their own workout history.

Make sure the app supports things like runs, cycling, etc. 