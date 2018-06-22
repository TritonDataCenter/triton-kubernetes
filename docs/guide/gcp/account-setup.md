## Setting up a GCP account

This document contains instructions for setting up an account on Google Cloud Platform, and preparing it for the creation of a Kubernetes environment using `triton-kubernetes`.

1. **Sign up for an GCP account** on https://cloud.google.com/free/.

    You need to enter a valid credit card to complete the Sign Up process.  After signing up, you will get a “Free Trial” subscription. With this subscription, you are restricted to a limited number of resources.  Most importantly, you have a quota of 8 vCPU’s per zone.

2. **Log in to GCP Console**: https://console.cloud.google.com.

3. **Find the project ID**: After logging into the GCP Console for the first time, you should already have a project called “My First Project”.  On the Dashboard page, find the “Project info” mini-window.  You should see the project name, project ID and project number for “My First Project”.  Copy down the value for project ID.

4. **Enable the Compute Engine service**: After you log in to the GCP Console as a new user, you must open the “Compute Engine” module for the first time in order for it to be configured for use.  On GCP Console home page, select the  “Compute > Compute Engine” menu.  You will notice that the system is being prepared, and that the Create button for VM Instances is still disabled.  Once the Compute Engine service is ready for use, the Create button is automatically enabled.

5. **Create a GCP Account file**: In order for Terraform to access the GCP Compute Engine service via API, it needs to read a JSON file that contains your service account key (i.e. your credential).  To create such a JSON file:

    1. On the GCP Console home page, select the “APIs & services” menu.
    2. Select the “Credentials” menu.
    3. Select the “Create credentials” dropdown list, and pick “service account key”.
    4. In the “Service account” field, select the value “Compute Engine default service account”.  Select “JSON” for Key type.  Then, click the “Create” button.
    5. A file called `My First Project-xxxxxxxx.json`, which contains your service account key, is automatically downloaded to your laptop.  After the file is downloaded, copy it to your `~/.ssh` folder.
    6. Rename the file by removing the space characters, so that the filename will look like `MyFirstProject-xxxxxxxx.json`.  (Note: This is because terraform will have problem reading this file if its filename contains spaces.)