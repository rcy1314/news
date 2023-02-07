package main

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"sort"
	"sync"
	"text/template"
	"time"

	"github.com/mmcdole/gofeed"
)

var (
	wg sync.WaitGroup
)

type TemplateData struct {
	Posts []*Post
}

type Post struct {
	Link      string
	Title     string
	Published time.Time
	Host      string
}

var (
	feeds = []string{
		"https://www.v2ex.com/feed/create.xml",
		"https://www.52pojie.cn/forum.php?mod=rss&fid=16",
		"https://rss.chyroc.cn/xml/weibo/user/origin/1088413295.xml",
		"https://jpsmile.com/feed",
		"http://www.ypojie.com/feed",
		"https://rsshub.uneasy.win/telegram/channel/res_share",
		"https://rsshub.uneasy.win/telegram/channel/dedao2019",
		"http://fetchrss.com/rss/612b7634c2c0385b1237705261893265f25af61a1259b542.xml",
		"https://www.freeaday.com/feed",
		"https://bbs.tampermonkey.net.cn/forum.php?mod=rss&fid=2&auth=0",
		"https://feeds.appinn.com/appinns/",
		"http://bbs.16xx8.com/forum.php?mod=rss&fid=134&auth=0",
		"https://www.jkmeng.cn/soft/feed",
		"https://rssnoise.vercel.app/bilibili/user/dynamic/20423027",
		"https://rssnoise.vercel.app/bilibili/user/dynamic/20166755",
		"https://rssnoise.vercel.app/bilibili/user/dynamic/142853317",
		"https://rssnoise.vercel.app/bilibili/user/dynamic/315819794",
		"https://blog.daliansky.net/atom.xml",
		"https://rssnoise.vercel.app/weibo/user/1112829033",
		"https://rssnoise.vercel.app/weibo/user/5811552055",
		"https://www.jkg.tw/index.xml",
		"https://rssnoise.vercel.app/weibo/user/5732021189",
	

		//  💖
		"https://rsshub.uneasy.win/jike/topic/55d81b4b60b296e5679785de",
		"https://rcy1314.github.io/Rss-Translation/rss/producthunt_today.xml",
		"https://rcy1314.github.io/Rss-Translation/rss/reddit_OpenAI.xml",
		"https://rcy1314.github.io/Rss-Translation/rss/reddit_youtubeaudiolibrary.xml",
		"https://rcy1314.github.io/Rss-Translation/rss/reddit_software.xml",
		"https://creatorsdaily.com/api/rss",

	}

	// Show up to 60 days of posts
	relevantDuration = 60 * 24 * time.Hour

	outputDir  = "docs" // So we can host the site on GitHub Pages
	outputFile = "index.html"

	// Error out if fetching feeds takes longer than a minute
	timeout = time.Minute
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	posts := getAllPosts(ctx, feeds)

	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return err
	}

	f, err := os.Create(path.Join(outputDir, outputFile))
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := &TemplateData{
		Posts: posts,
	}

	if err := executeTemplate(f, templateData); err != nil {
		return err
	}

	return nil
}

// getAllPosts returns all posts from all feeds from the last `relevantDuration`
// time period. Posts are sorted chronologically descending.
func getAllPosts(ctx context.Context, feeds []string) []*Post {
	postChan := make(chan *Post)

	wg.Add(len(feeds))
	for _, feed := range feeds {
		go getPosts(ctx, feed, postChan)
	}

	var posts []*Post
	go func() {
		for post := range postChan {
			posts = append(posts, post)
		}
	}()

	wg.Wait()
	close(postChan)

	// Sort items chronologically descending
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Published.After(posts[j].Published)
	})

	return posts
}

func getPosts(ctx context.Context, feedURL string, posts chan *Post) {
	defer wg.Done()
	parser := gofeed.NewParser()
	feed, err := parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		log.Println(err)
		return
	}

	for _, item := range feed.Items {
		published := item.PublishedParsed
		if published == nil {
			published = item.UpdatedParsed
		}
		if published.Before(time.Now().Add(-relevantDuration)) {
			continue
		}
		parsedLink, err := url.Parse(item.Link)
		if err != nil {
			log.Println(err)
		}
		post := &Post{
			Link:      item.Link,
			Title:     item.Title,
			Published: *published,
			Host:      parsedLink.Host,
		}
		posts <- post
	}
}

func executeTemplate(writer io.Writer, templateData *TemplateData) error {
	htmlTemplate := `
<!DOCTYPE html>
<html>
	<head>
	<link rel="icon" type="image/ico" href="https://cdn.staticaly.com/gh/rcy1314/tuchuang@main/NV/Level_Up_Your_Faith!_-_Geeks_Under_Grace.1yc7qyib5tsw.png">
    <link rel="stylesheet" href="https://cdn.staticfile.org/twitter-bootstrap/4.4.1/css/bootstrap.min.css">
    <link rel="stylesheet" href="https://cdn.staticfile.org/font-awesome/5.12.1/css/all.min.css">
	<link rel="stylesheet" href="ind.css">
    <link rel="stylesheet" href="style.css">
    <link rel="stylesheet" href="APlayer.min.css">
		<meta charset="utf-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>NOISE | 聚合信息阅读</title>
		<style>
		@import url("https://fonts.googleapis.com/css2?family=Nanum+Myeongjo&display=swap");

		body {
			font-family: "Nanum Myeongjo", serif;
			line-height: 1.7;
			max-width: 800px;
			margin:  auto ;
			padding: auto;
			height: 100%;
		}

		li {
			padding-bottom: 16px;
		}
	</style>
	</head>
	<script type='text/javascript' src='js/jquery-3.2.1.js'></script>  
        <script type='text/javascript'>  
            //显隐按钮  
            function showReposBtn(){  
                var clientHeight = $(window).height();  
                var scrollTop = $(document).scrollTop();  
                var maxScroll = $(document).height() - clientHeight;  
                //滚动距离超过可视一屏的距离时显示返回顶部按钮  
                if( scrollTop > clientHeight ){  
                    $('#retopbtn').show();  
                }else{  
                    $('#retopbtn').hide();  
                }  
                //滚动距离到达最底部时隐藏返回底部按钮  
                if( scrollTop >= maxScroll ){  
                    $('#rebtmbtn').hide();  
                }else{  
                    $('#rebtmbtn').show();  
                }  
            }  
              
            window.onload = function(){  
                //获取文档对象  
                $body = (window.opera) ? (document.compatMode == "CSS1Compat" ? $("html") : $("body")) : $("html,body");  
                //显示按钮  
                showReposBtn();  
            }  
              
            window.onscroll = function(){  
                //滚动时调整按钮显隐  
                showReposBtn();  
            }  
              
            //返回顶部  
            function returnTop(){  
                $body.animate({scrollTop: 0},400);  
            }  
              
            //返回底部  
            function returnBottom(){  
                $body.animate({scrollTop: $(document).height()},400);  
            }  
        </script>  
        <style type='text/css'>  
            #retopbtn{  
                position:fixed;  
                bottom:10px;  
                right:10px;  
            }  
            #rebtmbtn{  
                position:fixed;  
                top:10px;  
                right:10px;  
            }  
        </style>  
    </head>  
    <body>  
        <button id='rebtmbtn' onclick='returnBottom()'>⬇</button>  
		<button id='retopbtn' onclick='returnTop()'>⬆</button> 
	<body>


	    
	<div class="row my-card justify-content-center">
           
	<div class="col-lg-0 card">

	<!-- 上下翻转文字 -->
      
	<style type="text/css">#container-box-1{color:#526372;text-transform:uppercase;width:100%;font-size:16px;line-height:50px;text-align:center}#flip-box-1{overflow:hidden;height:50px}#flip-box-1 div{height:50px}#flip-box-1>div>div{color:#fff;display:inline-block;text-align:center;height:50px;width:100%}#flip-box-1 div:first-child{animation:show 20s linear infinite}.flip-box-1-1{background-color:#FF7E40}.flip-box-1-2{background-color:#C166FF}.flip-box-1-3{background-color:#737373}.flip-box-1-4{background-color:#4ec7f3}.flip-box-1-5{background-color:#42c58a}.flip-box-1-6{background-color:#F1617D}@keyframes show{0%{margin-top:-300px}5%{margin-top:-250px}16.666%{margin-top:-250px}21.666%{margin-top:-200px}33.332%{margin-top:-200px}38.332%{margin-top:-150px}49.998%{margin-top:-150px}54.998%{margin-top:-100px}66.664%{margin-top:-100px}71.664%{margin-top:-50px}83.33%{margin-top:-50px}88.33%{margin-top:0px}99.996%{margin-top:0px}100%{margin-top:300px}}</style>
	<div class="card card-site-info ">
	<div id="container-box-1">
	<div id="flip-box-1">
	<div><div class="flip-box-1-1"><i class="fa fa-gitlab" aria-hidden="true"></i>  rss feed for you </div></div>
	<div><div class="flip-box-1-2"><i class="fa fa-heart" aria-hidden="true"></i>  我们很年轻，但我们有信念、有梦想</div></div>
	<div><div class="flip-box-1-3"><i class="fa fa-gratipay" aria-hidden="true"></i>支持你的总会支持你，不支持的做再多也徒劳</div></div>
	<div><div class="flip-box-1-4"><i class="fa fa-drupal" aria-hidden="true"></i>  做这个世界的逆行者，先人一步看未来</div></div>
	<div><div class="flip-box-1-5"><i class="fa fa-gitlab" aria-hidden="true"></i>  只要你用心留意，世界将无比精彩</div></div>
	<div><div class="flip-box-1-6"><i class="fa fa-moon-o" aria-hidden="true"></i>  以下是信息聚合，精选各大站内容</div></div>
	<div><div class="flip-box-1-1">感谢原创者，感谢分享者，感谢值得尊重的每一位</div></div>
	</div>
	</div>
	</div>

			   <center>信息聚合阅读-RSS feed</center>
		
		<!-- 滚动代码-->

		<div class="card card-site-info ">
		<div class="m-3">
		<marquee scrollamount="5" behavior="right">
   
		<div id="blink">
   
		<a href="https://morss.it/:proxy:items=%7C%7C*[class=card]%7C%7Col%7Cli/https://rcy1314.github.io/news/">📢：rss feed for you 🔛</a>Rss聚合阅读页 🎁</div> 
   
   
		<script language="javascript"> 
   
   function changeColor(){ 
   
   var color="#f00|#0f0|#00f|#880|#808|#088|yellow|green|blue|gray"; 
   
   color=color.split("|"); 
   
   document.getElementById("blink").style.color=color[parseInt(Math.random() * color.length)]; 
   
   } 
   
   setInterval("changeColor()",200); 
   
		</script>
   
		</marquee>
		</div>
		</div>
   
   
		<!-- 向右流动代码-->
   
		<marquee scrollamount="3" direction="right" behavior="alternate">
   
		<a>😄😃😀</a>
   
		</marquee>
   
   
		
   
   
		<div class="alert alert-danger alert-dismissable">
		<button type="button" class="close" data-dismiss="alert"
			   aria-hidden="true">
		   &times;
		</button>
		 页面自动2小时监测更新一次！
		</div>
   
	<!-- 音乐 -->
	</script> 		  
	<div id="aplayer" class="aplayer" data-order="random" data-id="128460001" data-server="netease" data-type="playlist" data-fixed="true" data-autoplay="false" data-volume="0.8"></div>
	<!-- aplayer -->
	<script src="https://cdn.staticfile.org/jquery/3.2.1/jquery.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/aplayer@1.10.1/dist/APlayer.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/meting@1.2.0/dist/Meting.min.js"></script>
	<!-- end_aplayer -->
	<script src="https://cdn.staticfile.org/popper.js/1.15.0/umd/popper.min.js"></script>
	<script defer src="https://cdn.staticfile.org/twitter-bootstrap/4.4.1/js/bootstrap.min.js"></script>
	<script src="https://cdn.jsdelivr.net/gh/kaygb/kaygb@master/layer/layer.js"></script>
	<script src="https://cdn.jsdelivr.net/gh/kaygb/kaygb@master/js/v3.js"></script>
   
		<!-- 站长说 -->
   
		<div class="card card-site-info ">
		<div class="m-3">
		   <div class="small line-height-2"><b>公告 ： <i class="fa fa-volume-down fa-2" aria-hidden="true"></i></b></li><?php /*echo $conf['announcement'];*/?>  增加reddit、producthunt等外网源，阅读请使用右键打开新链接，如需添加其它feed请点击页面最下方。</div>
		</div>
		 </div>
   
   
		<!-- 广告招租-->
		<div class="card card-site-info ">
		<div class="m-3">
		   <div class="small line-height-2"><b>广告位 <i class="fa fa-volume-down fa-2" aria-hidden="true"></i></b></li>：<?php /*echo $conf['announcement'];*/?>
		<a href="https://efficiencyfollow.notion.site">Efficiency主页</a>&nbsp;&nbsp;&nbsp; 
		<a href="https://noisedh.cn">Noise导航站</a>&nbsp;&nbsp;&nbsp;
		<a href="https://t.me/quanshoulu">TG发布频道</a>&nbsp;&nbsp;&nbsp;
		<a href="https://noisework.cn">引导主页</a>&nbsp;&nbsp;&nbsp;
		<a href="https://www.noisesite.cn">知识效率集</a>&nbsp;&nbsp;&nbsp;
		<a href="https://rcy1314.github.io/some-stars">我的star列表</a>&nbsp;&nbsp;&nbsp;
		<a href="https://noiseyp.top">Noise资源库</a></div>
		</div>
			<br>
	   

		<ol>
			{{ range .Posts }}<li><a href="{{ .Link }}">{{ .Title }}</a> ({{ .Host }})</li>
			{{ end }}
		</ol>

		<footer>
		<div class="text-center py-1">   
        <div>
         <div class="text-center py-1">   
         <div>
		 <a href="https://ppnoise.notion.site/wiki-1ba2367142dc4b80b24873120a96efb5" target="_blank" rel="nofollow noopener">
	     <span>feed添加</span></a>    <br>
         </div>
	     <a href="https://noisework.cn" target="_blank" rel="nofollow noopener">
	     <span>主页</span></a>    <br>
         </div>
		 <script async src="//busuanzi.ibruce.info/busuanzi/2.3/busuanzi.pure.mini.js"></script>
		 <span id="busuanzi_container_site_pv" style='display:none'> 本站总访问量<span id="busuanzi_value_site_pv"></span>次</span>
		 </div>	
		 <div style="margin-top: 10px;">
		 &nbsp; 
		<span id="momk"></span>
		<span id="momk" style="color: #ff0000;"></span> 
		<script type="text/javascript">
   function NewDate(str) {
   str = str.split('-');
   var date = new Date();
   date.setUTCFullYear(str[0], str[1] - 1, str[2]);
   date.setUTCHours(0, 0, 0, 0);
   return date;
   }
   function momxc() {
   var birthDay =NewDate("2021-09-23");
   var today=new Date();
   var timeold=today.getTime()-birthDay.getTime();
   var sectimeold=timeold/1000
   var secondsold=Math.floor(sectimeold);
   var msPerDay=24*60*60*1000; var e_daysold=timeold/msPerDay;
   var daysold=Math.floor(e_daysold);
   var e_hrsold=(daysold-e_daysold)*-24;
   var hrsold=Math.floor(e_hrsold);
   var e_minsold=(hrsold-e_hrsold)*-60;
   var minsold=Math.floor((hrsold-e_hrsold)*-60); var seconds=Math.floor((minsold-e_minsold)*-60).toString();
   document.getElementById("momk").innerHTML = "本站已运行:"+daysold+"天"+hrsold+"小时"+minsold+"分"+seconds+"秒";
   setTimeout(momxc, 1000);
   }momxc();
	</footer>
</body>
</html>
`

	tmpl, err := template.New("webpage").Parse(htmlTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(writer, templateData); err != nil {
		return err
	}

	return nil
}
