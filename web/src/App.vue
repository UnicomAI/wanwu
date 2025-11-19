<template>
  <div id="app">
    <router-view></router-view>
  </div>
</template>

<script>
    export default {
        name: 'App',
        data(){
            return {}
        },
    }
</script>

<style lang="scss">
*{
  margin: 0;
  padding: 0;
  font-size: 13px;
  box-sizing: border-box;
  -webkit-font-smoothing: antialiased; /*chrome、safari*/
  -moz-osx-font-smoothing: grayscale; /*firefox*/
}
body,html{
  width: 100%;
  height: 100%;
  /*overflow: hidden;*/
  /*font-family:'Microsoft YaHei', Arial, sans-serif;*/
  font-family: 'Source Han Sans SC', 'Segoe UI';
}

/* Fix Dialog appearing with Table left shift issue */
body.el-popup-parent--hidden {
  padding-right: 0 !important;
}

/*::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-button {
  display: none;
}
::-webkit-scrollbar-corner {
  background-color: transparent;
}
::-webkit-scrollbar-thumb {
  background-color: #bdbebe;
  min-height: 50px;
  min-width: 50px;
  -webkit-border-radius: 10px;
  border-radius: 10px;
  -webkit-transition: background-color 0.2s;
  -o-transition: background-color 0.2s;
  transition: background-color 0.2s;
}
::-webkit-scrollbar-track-piece {
  background-color: #eee;
}
::-webkit-scrollbar-track {
  -webkit-border-radius: 10px;
  border-radius: 10px;
}*/
#app {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  --main-title-hei: 100px;
  --main-padding: 0 75px;
  --box-shadow:1px 1px 12px #f1f1f1;
  width: 100%;
  height:100%;
}
.flex{
  display: flex;
  max-width: 100%;
}
.rl{
  position: relative;
}
.abs{
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
}
/* Set height of div created by Qiankun */
div[data-name]{
  height: 100%;
}
.routerview-container,.overview{
  z-index: 100;
}
.routerview-container{
  .common-module-wrap>.title{
    margin-bottom: 20px;
  }
}
.page-wrapper {
  padding: 0 10px;
}
.page-title {
  font-size: 18px;
  font-weight: bold;
  text-align: left;
  color: $color_title;
  padding: 24px 0 14px 30px;
  border-bottom: 1px solid #eaeaea;
  .page-title-img {
    width: 18px;
    display: inline-block;
    vertical-align: middle;
    margin-right: 10px;
  }
  .page-title-name {
    display: inline-block;
    vertical-align: middle;
    font-size: 18px;
    font-weight: bold;
    text-align: left;
    color: $color_title;
  }
}
.primary-bt{
  min-width: 76px;
  border-radius: 23px;
  border: 1px solid $color;
  background-color: $color;
  color: #fff;
  height: 30px;
  line-height: 100%;
  text-align: center;
  font-size: 12px;
  padding: 8px 15px;
  font-weight: 400;
  cursor: pointer;
}

/* Default height auto-adapt */
.main-content{
  height: auto;
  min-height: 100%;
  margin: auto;
  position: relative;
  background-color: #fff;
  padding: var(--main-padding);
}
/* Full screen display */
.full-content{
  height: 100%;
}
.md-list{
  .md-list__title{
    height: var(--main-title-hei);
    line-height: var(--main-title-hei);
    font-weight: bold;
    background-color: #fff;
    display: flex;
    .__title-left{
      font-size: 18px;
      flex: 1;
    }
    .__title-right{
      flex:1;
      text-align: right;
      .tip{
        color: #999;
        font-weight: 400;
        font-size: 13px;
      }
      .compare-bt{
        margin-left: 10px;
        span {
          font-size: 13px;
        }
      }
    }
  }
  .md-list__list{
    min-height: calc(100% - var(--main-title-hei));
    margin: auto;
    background-color: #fff;
    .list-item{
      position: relative;
      width: calc((100% - 60px) / 3);
      height: 240px;
      display: inline-block;
      margin: 0 30px 30px 0;
      padding: 25px;
      border: 1px solid #EBEBEB;
      border-radius: 8px;
      vertical-align: top;

      .item__header_checkbox{
        position: absolute;
        right: 10px;
        top:5px;
        padding: 8px;
      }
      .list-item__header{
        line-height: 30px;
        .item__header_img{
          width: 26px;
          height: 30px;
          object-fit: contain;
        }
        .item__header_name{
          flex:20;
          font-weight: bold;
          margin-left: 12px;
          cursor: pointer;
        }
      }
      .list-item__info{
        margin-top: 5px;
        .info{
          line-height: 30px;
          overflow:hidden;
          white-space: nowrap;
          text-overflow: ellipsis;

          span{
            font-size: 12px;
          }
          span:nth-child(1){
            color: #999;
            display: inline-block;
            width: 70px;
          }
        }
      }
    }
    .list-item-active{
      border: 1px solid $color;
    }
    .no-rt-bd{
      margin: 0 0 20px 0!important;
    }
  }
}
.copy-icon{
  font-size: 15px;
  padding: 4px;
  color: #33a4df;
  cursor: pointer;
}
</style>