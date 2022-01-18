# **SchedulX**

SchedulX is a cloud-native service orchestration and deployment solution based on the open-source BridgX project, with the goal of enabling developers to orchestrate and deploy services on computing resources obtained from BridgX.
It has the following key features:
1. Able to combine dynamic scale-up and scale-down characteristics for service deployment;
2. Manage the service operations on different cloud platforms on a single unified platform;
3. Simple and easy to use, easy to get started.



Installation and Deployment
--------
1. Configuration Requirements
For stable system operation, the recommended system model is 2 processing cores with 4G RAM; for Linux systems as well as macOS systems, ensure that SchedulX has been installed and tested.

2. Environmental Dependence
SchedulX relies on BridgX, so please follow the [BridgX Installation Guide](https://github.com/galaxy-future/bridgx/blob/dev/docs/EN-README.md) for installation. Requires an intranet deployment environment that can connect to the cloud vendor's VPC.



3.Installation Steps

* (1)Download source code
  - back-end project：
  > `git clone git@github.com:galaxy-future/schedulx.git`
  - After downloading the code, modify the configuration file `register/conf/config.yml`，and fill in the Accesskey, Secret and Region of the cloud account.

* (2)macOS System Deployment
  - For back-end deployment, run in the SchedulX directory:
    > `make docker-run-mac`

* (3)Linux Installation and Deployment
  - 1）For users
    - For back-end deployment, run in the SchedulX directory:
      > `make docker-run-linux`
    - When the system is running, type http://127.0.0.1 into your browser to see the management console interface, with the default username root and password 123456.


  - 2）For developers
    - Back-end deployment
      - SchedulX depends on the MySQL component.
           - If you are using the built-in MySQL, run the following command in the SchedulX root directory:
             > `docker-compose up -d`    //Start SchedulX <br>
             > `docker-compose down`    //Stop SchedulX  <br>
           - If you already have an external MySQL service, you can go to `cd conf` to change the corresponding IP and port configuration information, then go to the root directory of SchedulX and use the following command:

             > `docker-compose up -d schedulx`   //Start the SchedulX service <br>
             > `docker-compose down`     //stop the SchedulX service

4.Front-end Interface Operation

If you need to use the web-based front-end to perform any operations, please download and install [ComandX](https://github.com/galaxy-future/comandx/blob/main/docs/EN-README.md).

Code of Conduct
------
[Contributor Convention](https://github.com/galaxy-future/schedulx/blob/master/CODE_OF_CONDUCT.md)

Authorization
-----
SchedulX is licensed under the [Apache 2.0 license](https://github.com/galaxy-future/schedulx/blob/master/README.md) agreement.


