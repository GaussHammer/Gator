Gator is a RSS feed aggregator. To use this app you must have PostGresql and Go installed. 
the default url for the postgres database is using postgres as a user and postgres as password. 
To change this url to something that'll suit you, you can simply edit the dbUrl variable at the top of thhe main() function in the main.go file.
You also need to create a .json config file in your home directory, the file must be named ".gatorconfig.json"
Once you have properly connected the database and set up the config file, you can run $go install . from the root of the /gator directory to install the program.

here are the commands usable with gator:
$gator register *name*      :register a new user of the name *name*
$gator login *name*         :login as the named user
$gator reset                :wipes the entire database, use this at your own risk
$gator users                :lists all the users registered
$gator addfeed "title" "url":add a new rss feed to the database
$gator feeds                :returns all existing feeds registered
$gator follow "url"         :makes the current user follow an already existing feed
$gator following            :returns the list of feeds followed by the current user
$gator unfollow "url"       :removes the given feed from the current user's list of followed feeds
$gator agg *timestamp*      :make the program start aggregating posts from the feeds every *timestamp* (could be 1s, 1m, 1h) stop it with Control+C
$gator browse X             :browses the X firsts posts from the current user's followed feeds. Default limit is 2 if no parameter is given
