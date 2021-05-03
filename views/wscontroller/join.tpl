<!DOCTYPE html>

<html>
<head>
  <title>{{.room}}号房间-测试聊天室</title>
</head>
<body>
<script src="static/js/jquery-3.2.1.min.js" type="text/javascript"></script>
<script type="text/javascript">
$(function(){
    var conn;
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + '/ws' + document.location.search);
        conn.onclose = function (evt) {
            $('.close').html("websocket connect close~"); 
        };
        conn.onmessage = function (evt) {
            console.log(evt.data)
            if (evt.data =="heartbeat") {
                return
            }
            var messages = evt.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
				if (messages[i] == "heartbeat") {
				    continue
				}
				message = $.parseJSON(messages[i]);
				if (message.type == "1") {
				    tmpData = "系统消息：" + "：" + message.message
				} else {
				    tmpData = message.username + "说：" + message.message
				}
                $('.messages').append(tmpData + "<br/>");
            }
        };
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
    $('.submit').click(function(){
        if (!conn) {
            return false;
        }    
        if(!$('.msg').val()){
            return false;
        }
        conn.send($('.msg').val())
        $('.msg').val('');
        return false;
    })
});
</script>
<style>
    .messages{width:100%;min-height:200px;border:1px solid #ccc;}
</style>
<div>{{.username}}  您所在的房间号：{{.room}}号房间</div>
</br></br>
<div class="close"></div>
<div class="messages"></div>
</br></br>
输入信息：<input text="text" id="msg" class="msg" />
</br></br>
<input type="button" value="提交信息" class="submit" />
</body>
</html>
