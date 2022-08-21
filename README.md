# dota2-camera-fixer
### Description
Simple tool to fix default DOTA 2 camera.
There are 2 files (excluding executable) needed for application to work properly:
- *camera-values.json* - contains camera attributes with its actual values and replace values, feel free to change value of 'newValue' fields to get 
results that will satisfy you;
- *config.ini* - file containing path to Steam directory and other user preferences that you can change, you must replace the path inside to your actual path to Steam directory. **Do not change anyting in "AppSettings" section!**
### How to use
0. Check game files integrity via Steam to make sure your *client.dll* is not modified;
1. Open *config.ini* and change value of "SteamDirPath" field to path to your Steam directory, for example *C:\Program Files (x86)\Steam* for Windows,
or */home/Steam* for Linux;
2. Open *camera-values.json* and change 'newValue' field values to more suitable for you, but keep in mind that **all values are integers** and:
    - *dota_camera_distance_min* must be **less than 10000**
    - *dota_camera_fov_min* must be **less than 90**
    - *dota_camera_pitch_min* must be **less than 90**
    - *dota_camera_distance* must be **less than 100**
    - *sv_noclipaccelerate* must be **less than 10000**
3. Run *dota2-camera-fixer* executable and wait until message 'Done' is printed.
4. Run DOTA 2 and enjoy your fixed camera.
5. In case something goes wrong (e.g. game crashes just after being launched):
    - You got any errors in terminal output: if you don't have a single thought why that error occures and how to fix it, 
    make an issue on this repo, we'll try to figure it out together;   
    - Game crashes on stratup: make sure the application was run with **unmodified** *client.dll* file and try to run it again;
    - Game still crashes on startup after previous step: go to *<PATH_TO_YOUR_STEAM_DIR>\steamapps\common\dota 2 beta\game\bin\win64* directory, 
    here you'll see folder *client_dll_backup*, open this folder and copy-paste *client.dll* file contained here to previous directory;
### User preferences (*config.ini*):
- *SteamDirPath* - path to your Steam directory, change it to yours;
- *ShowLogInfo* - change to **false** if you do not want to see info logs (error logs and "Done" message will still be shown);
- *BackupDirName* - name of the backup directory, you can change it to whatever name you like;
- *CameraValuesFileName* - change this value only if you have renamed "camera-values.json" file.
### How to install
#### The simplest way:
Just download the archive attached to the release and extract the application folder somewhere you good with.  
**DO NOT PLACE APPLICATION FILES FROM ARCHIVIED FOLDER TO SEPARATE LOCATIONS, ALL THESE FILES MUST BE LOCATED TOGETHER IN THE SAME FOLDER!!!**
#### The hardest way:
1. Download and install [Golang](https://go.dev/doc/install).
2. Clone this repo to your local machine.
3. Open repo directory in terminal.
4. Execute *go build -o dota2-camera-fixer* command.
5. Now you can feel yourself a very skilled Golang developer and run created executable file just from this folder.
