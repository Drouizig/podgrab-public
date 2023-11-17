# Podgrab public

This is a heacy modification of the original podgrab source, to make it into a public podcast aggregator.
Added a translation in french, galo and breton for the public part, and separated the settings/mangement part.


### Modifications
- Put all the settings stuff in /admin location, and only protect this part behind basic auth
- Remove the downloading of episodes
- Internationalized the public part and added translations in french, breton and galo
- Added a subscribe via antennapod button
- Added a copy rss url button
- Added podcast author mention
- Added podcast link in episode list
- added dockerfile

### Using Docker

Download the repo and run :
```
docker-compose build
docker-compose up -d
```

Same env vars as original project

| Name            | Description                                                             | Default |
| --------------- | ----------------------------------------------------------------------- | ------- |
| CHECK_FREQUENCY | How frequently to check for new episodes and missing files (in minutes) | 30      |
| PASSWORD        | Set to some non empty value to enable Basic Authentication, username `podgrab`|(empty) NOW ONLY FOR ADMIN PART|
| PORT            | Change the internal port of the application. If you change this you might have to change your docker configuration as well | (empty) |  


## Utils

How to delete orphan podcast items :

```
DELETE podcast orphan item
delete from podcast_items where podcast_items.podcast_id not in (select podcasts.id from podcasts, podcast_items where podcasts.id = podcast_items.podcast_id);
```

## License

Distributed under the GPL-3.0 License. See `LICENSE` for more information.