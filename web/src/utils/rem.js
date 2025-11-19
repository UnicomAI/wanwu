(function(doc, win) {
    var docEl = doc.documentElement,
        timer = '';

  //当domLoad完成 when ，OR者 屏幕垂直、水平方向Has改变进行html of 根元素计算
  function recalc(){
        var clientWidth = docEl.getBoundingClientRect ? docEl.getBoundingClientRect().width : docEl.clientWidth;
	    if (!clientWidth) return;
        var fontSize = parseFloat(100 * (clientWidth / 1920));
        docEl.style.fontSize = fontSize + 'px';
        sessionStorage.setItem('fs',fontSize);
    }
    win.addEventListener("pageshow",function(e){
      if (e.persisted){
        clearTimeout(timer);
        timer = setTimeout(recalc,30);
      }
    },false);
    recalc();

    function handleFontSize() {
      WeixinJSBridge.invoke('setFontSizeCallback', { 'fontSize' : 0 });
      WeixinJSBridge.on('menu:setfont', function() {
        WeixinJSBridge.invoke('setFontSizeCallback', { 'fontSize' : 0 });
      });
    }
    //不受微信字体大小影响
    if (typeof WeixinJSBridge == "object" && typeof WeixinJSBridge.invoke == "function") {
      handleFontSize();
    } else {
      if (document.addEventListener) {
        document.addEventListener("WeixinJSBridgeReady", handleFontSize, false);
      } else if (document.attachEvent) {
        document.attachEvent("WeixinJSBridgeReady", handleFontSize);
        document.attachEvent("onWeixinJSBridgeReady", handleFontSize);
      }
    }

})(document, window);
