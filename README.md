# ApnicCN
解析Apnic分配给大陆的IP段，利用腾讯云SCF上传至腾讯云COS

# 为什么写这个工具
公司有个内部项目要判断IP地址是否为国内IP

数据来源为："http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest"

服务器地址在国外，响应大小接近 4mb

由于仅需要国内IP地址分配范围，故需要预解析

筛选出国内的数据仅 8000 多行，大小为 130kb

上传至 腾讯云 cos 便于以后加载

国内服务器加载源文件比较慢，懒得专门开一台国外服务器，体验了 腾讯云 SCF 选择新加坡

选择定时器每天触发一次，整体体验不错，6000ms 就能处理完
