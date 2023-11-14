package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/akhilrex/podgrab/controllers"
	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/service"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/text/language"
)

func main() {
	var err error
	db.DB, err = db.Init()
	if err != nil {
		fmt.Println("statuse: ", err)
	} else {
		db.Migrate()
	}
	r := gin.Default()

	r.Use(setupSettings())
	r.Use(gin.Recovery())
	r.Use(location.Default())

	funcMap := template.FuncMap{
		"intRange": func(start, end int) []int {
			n := end - start + 1
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = start + i
			}
			return result
		},
		"removeStartingSlash": func(raw string) string {
			fmt.Println(raw)
			if string(raw[0]) == "/" {
				return raw
			}
			return "/" + raw
		},
		"isDateNull": func(raw time.Time) bool {
			return raw == (time.Time{})
		},
		"formatDate": func(raw time.Time) string {
			if raw == (time.Time{}) {
				return ""
			}

			return raw.Format("02/01/2006")
		},
		"naturalDate": func(raw time.Time) string {
			return service.NatualTime(time.Now(), raw)
			//return raw.Format("Jan 2 2006")
		},
		"latestEpisodeDate": func(podcastItems []db.PodcastItem) string {
			var latest time.Time
			for _, item := range podcastItems {
				if item.PubDate.After(latest) {
					latest = item.PubDate
				}
			}
			return latest.Format("02/01/2006")
		},
		"downloadedEpisodes": func(podcastItems []db.PodcastItem) int {
			count := 0
			for _, item := range podcastItems {
				if item.DownloadStatus == db.Downloaded {
					count++
				}
			}
			return count
		},
		"downloadingEpisodes": func(podcastItems []db.PodcastItem) int {
			count := 0
			for _, item := range podcastItems {
				if item.DownloadStatus == db.NotDownloaded {
					count++
				}
			}
			return count
		},
		"formatFileSize": func(inputSize int64) string {
			size := float64(inputSize)
			const divisor float64 = 1024
			if size < divisor {
				return fmt.Sprintf("%.0f bytes", size)
			}
			size = size / divisor
			if size < divisor {
				return fmt.Sprintf("%.2f KB", size)
			}
			size = size / divisor
			if size < divisor {
				return fmt.Sprintf("%.2f MB", size)
			}
			size = size / divisor
			if size < divisor {
				return fmt.Sprintf("%.2f GB", size)
			}
			size = size / divisor
			return fmt.Sprintf("%.2f TB", size)
		},
		"formatDuration": func(total int) string {
			if total <= 0 {
				return ""
			}
			mins := total / 60
			secs := total % 60
			hrs := 0
			if mins >= 60 {
				hrs = mins / 60
				mins = mins % 60
			}
			if hrs > 0 {
				return fmt.Sprintf("%02d:%02d:%02d", hrs, mins, secs)
			}
			return fmt.Sprintf("%02d:%02d", mins, secs)
		},
		"htmlDecode": func(htmlStr string) string {
			return html.UnescapeString(htmlStr)
		},
		"trans": func(code string, context *gin.Context) string {
			var message, err = ginI18n.GetMessage(context, code)
			if err != nil {
				return code
			}
			return message
		},
	}
	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob("client/*"))

	//r.LoadHTMLGlob("client/*")
	r.SetHTMLTemplate(tmpl)

	pass := os.Getenv("PASSWORD")
	var langRouter *gin.RouterGroup
	var router *gin.RouterGroup
	var adminRouter *gin.RouterGroup
	var assetsRouter *gin.RouterGroup

	if pass != "" {
		adminRouter = r.Group("/admin", gin.BasicAuth(gin.Accounts{
			"podgrab": pass,
		}))
	} else {
		adminRouter = r.Group("/admin", gin.BasicAuth(gin.Accounts{
			"podgrab": "podgrab",
		}))
	}
	langRouter = r.Group("/")
	assetsRouter = r.Group("/")
	router = r.Group("/:lang")

	router.Use(ginI18n.Localize(
		ginI18n.WithGetLngHandle(
			func(context *gin.Context, defaultLng string) string {
				lng := context.Param("lang")
				if lng == "" {
					return defaultLng
				}
				return lng
			},
		),
		ginI18n.WithBundle(&ginI18n.BundleCfg{
			RootPath:         "./translations",
			AcceptLanguage:   []language.Tag{language.Make("ga"), language.Make("fr"), language.Make("br")},
			DefaultLanguage:  language.Make("br"),
			UnmarshalFunc:    toml.Unmarshal,
			FormatBundleFile: "toml",
		}),
	))

	adminRouter.Use(ginI18n.Localize(
		ginI18n.WithBundle(&ginI18n.BundleCfg{
			RootPath:         "./translations",
			AcceptLanguage:   []language.Tag{language.Make("ga"), language.Make("fr"), language.Make("br")},
			DefaultLanguage:  language.Make("br"),
			UnmarshalFunc:    toml.Unmarshal,
			FormatBundleFile: "toml",
		}),
	))

	dataPath := os.Getenv("DATA")
	backupPath := path.Join(os.Getenv("CONFIG"), "backups")

	langRouter.GET("/", controllers.SelectLang)
	assetsRouter.Static("/webassets", "./webassets")
	assetsRouter.Static("/assets", dataPath)
	router.Static(backupPath, backupPath)
	adminRouter.POST("/podcasts", controllers.AddPodcast)
	router.GET("/podcasts", controllers.GetAllPodcasts)
	router.GET("/podcasts/:id", controllers.GetPodcastById)
	assetsRouter.GET("/podcasts/:id/image", controllers.GetPodcastImageById)
	adminRouter.DELETE("/podcasts/:id", controllers.DeletePodcastById)
	router.GET("/podcasts/:id/items", controllers.GetPodcastItemsByPodcastId)
	adminRouter.GET("/podcasts/:id/download", controllers.DownloadAllEpisodesByPodcastId)
	adminRouter.DELETE("/podcasts/:id/items", controllers.DeletePodcastEpisodesById)
	adminRouter.DELETE("/podcasts/:id/podcast", controllers.DeleteOnlyPodcastById)
	router.GET("/podcasts/:id/pause", controllers.PausePodcastById)
	router.GET("/podcasts/:id/unpause", controllers.UnpausePodcastById)
	router.GET("/podcasts/:id/rss", controllers.GetRssForPodcastById)

	router.GET("/podcastitems", controllers.GetAllPodcastItems)
	router.GET("/podcastitems/:id", controllers.GetPodcastItemById)
	assetsRouter.GET("/podcastitems/:id/image", controllers.GetPodcastItemImageById)
	assetsRouter.GET("/podcastitems/:id/file", controllers.GetPodcastItemFileById)
	adminRouter.GET("/podcastitems/:id/markUnplayed", controllers.MarkPodcastItemAsUnplayed)
	adminRouter.GET("/podcastitems/:id/markPlayed", controllers.MarkPodcastItemAsPlayed)
	adminRouter.GET("/podcastitems/:id/bookmark", controllers.BookmarkPodcastItem)
	adminRouter.GET("/podcastitems/:id/unbookmark", controllers.UnbookmarkPodcastItem)
	router.PATCH("/podcastitems/:id", controllers.PatchPodcastItemById)
	adminRouter.GET("/podcastitems/:id/download", controllers.DownloadPodcastItem)
	router.GET("/podcastitems/:id/delete", controllers.DeletePodcastItem)

	router.GET("/tags", controllers.GetAllTags)
	router.GET("/tags/:id", controllers.GetTagById)
	router.GET("/tags/:id/rss", controllers.GetRssForTagById)
	adminRouter.DELETE("/tags/:id", controllers.DeleteTagById)
	adminRouter.POST("/tags", controllers.AddTag)
	adminRouter.POST("/podcasts/:id/tags/:tagId", controllers.AddTagToPodcast)
	adminRouter.DELETE("/podcasts/:id/tags/:tagId", controllers.RemoveTagFromPodcast)

	adminRouter.GET("/add", controllers.AddPage)
	adminRouter.GET("/search", controllers.Search)
	router.GET("/", controllers.HomePage)
	adminRouter.GET("/", controllers.HomePageAdmin)
	router.GET("/podcasts/:id/view", controllers.PodcastPage)
	router.GET("/episodes", controllers.AllEpisodesPage)
	router.GET("/allTags", controllers.AllTagsPage)
	adminRouter.GET("/settings", controllers.SettingsPage)
	adminRouter.POST("/settings", controllers.UpdateSetting)
	adminRouter.GET("/backups", controllers.BackupsPage)
	adminRouter.POST("/opml", controllers.UploadOpml)
	adminRouter.GET("/opml", controllers.GetOmpl)
	// adminRouter.GET("/scrap", controllers.ScrapRadioBreizhPage)
	router.GET("/player", controllers.PlayerPage)
	router.GET("/rss", controllers.GetRss)

	r.GET("/ws", func(c *gin.Context) {
		controllers.Wshandler(c.Writer, c.Request)
	})
	go controllers.HandleWebsocketMessages()

	go assetEnv()
	go intiCron()

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
func setupSettings() gin.HandlerFunc {
	return func(c *gin.Context) {

		setting := db.GetOrCreateSetting()
		c.Set("setting", setting)
		c.Writer.Header().Set("X-Clacks-Overhead", "GNU Terry Pratchett")

		c.Next()
	}
}

func intiCron() {
	checkFrequency, err := strconv.Atoi(os.Getenv("CHECK_FREQUENCY"))
	if err != nil {
		checkFrequency = 3600
		log.Print(err)
	}
	service.UnlockMissedJobs()
	//gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.RefreshEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.CheckMissingFiles)
	gocron.Every(uint64(checkFrequency) * 2).Minutes().Do(service.UnlockMissedJobs)
	gocron.Every(uint64(checkFrequency) * 3).Minutes().Do(service.UpdateAllFileSizes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingImages)
	gocron.Every(2).Days().Do(service.CreateBackup)
	<-gocron.Start()
}

func assetEnv() {
	log.Println("Config Dir: ", os.Getenv("CONFIG"))
	log.Println("Assets Dir: ", os.Getenv("DATA"))
	log.Println("Check Frequency (mins): ", os.Getenv("CHECK_FREQUENCY"))
}
