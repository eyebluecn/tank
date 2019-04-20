#### PROPFIND 列出目录情况
request method
```
PROPFIND
```

request header
```
Authorization=Basic YWRtaW46YWRtaW4=
Content-Type=text/xml
Accept-Encoding=gzip
Depth=infinity
```

request body
```
<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:">
    <D:prop>
        <D:resourcetype />
        <D:getcontentlength />
        <D:creationdate />
        <D:getlastmodified />
    </D:prop>
</D:propfind>
```

response body
```
<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:">
    <D:response>
        <D:href>/api/dav/</D:href>
        <D:propstat>
            <D:prop>
                <D:displayname>dav</D:displayname>
                <D:getlastmodified>Tue, 16 Apr 2019 17:50:59 GMT</D:getlastmodified>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:resourcetype>
                    <D:collection xmlns:D="DAV:"/>
                </D:resourcetype>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/api/</D:href>
        <D:propstat>
            <D:prop>
                <D:displayname>api</D:displayname>
                <D:getlastmodified>Tue, 16 Apr 2019 17:51:03 GMT</D:getlastmodified>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:resourcetype>
                    <D:collection xmlns:D="DAV:"/>
                </D:resourcetype>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/api/dav/</D:href>
        <D:propstat>
            <D:prop>
                <D:displayname>dav</D:displayname>
                <D:getlastmodified>Tue, 16 Apr 2019 17:51:38 GMT</D:getlastmodified>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:resourcetype>
                    <D:collection xmlns:D="DAV:"/>
                </D:resourcetype>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/api/dav/body.txt</D:href>
        <D:propstat>
            <D:prop>
                <D:displayname>body.txt</D:displayname>
                <D:getlastmodified>Tue, 16 Apr 2019 17:51:38 GMT</D:getlastmodified>
                <D:getcontenttype>text/plain; charset=utf-8</D:getcontenttype>
                <D:getetag>"159605ccc1d0f3c410"</D:getetag>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:resourcetype></D:resourcetype>
                <D:getcontentlength>16</D:getcontentlength>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/api/dav/cat.txt</D:href>
        <D:propstat>
            <D:prop>
                <D:resourcetype></D:resourcetype>
                <D:getcontentlength>24</D:getcontentlength>
                <D:getlastmodified>Tue, 16 Apr 2019 17:51:19 GMT</D:getlastmodified>
                <D:getcontenttype>text/plain; charset=utf-8</D:getcontenttype>
                <D:getetag>"159605c862b0d64c18"</D:getetag>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:displayname>cat.txt</D:displayname>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/cat/</D:href>
        <D:propstat>
            <D:prop>
                <D:getlastmodified>Sat, 13 Apr 2019 16:55:54 GMT</D:getlastmodified>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:displayname>cat</D:displayname>
                <D:resourcetype>
                    <D:collection xmlns:D="DAV:"/>
                </D:resourcetype>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/cat/dog/</D:href>
        <D:propstat>
            <D:prop>
                <D:getlastmodified>Sat, 13 Apr 2019 16:55:58 GMT</D:getlastmodified>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:displayname>dog</D:displayname>
                <D:resourcetype>
                    <D:collection xmlns:D="DAV:"/>
                </D:resourcetype>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/cat/dog/pig/</D:href>
        <D:propstat>
            <D:prop>
                <D:resourcetype>
                    <D:collection xmlns:D="DAV:"/>
                </D:resourcetype>
                <D:getlastmodified>Sat, 13 Apr 2019 16:56:08 GMT</D:getlastmodified>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:displayname>pig</D:displayname>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/cat/dog/pig/hi.txt</D:href>
        <D:propstat>
            <D:prop>
                <D:getlastmodified>Sat, 13 Apr 2019 16:56:08 GMT</D:getlastmodified>
                <D:getcontenttype>text/plain; charset=utf-8</D:getcontenttype>
                <D:getetag>"15951707dc1116d87"</D:getetag>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
                <D:displayname>hi.txt</D:displayname>
                <D:resourcetype></D:resourcetype>
                <D:getcontentlength>7</D:getcontentlength>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
    <D:response>
        <D:href>/api/dav/morning.txt</D:href>
        <D:propstat>
            <D:prop>
                <D:resourcetype></D:resourcetype>
                <D:getcontentlength>13</D:getcontentlength>
                <D:displayname>morning.txt</D:displayname>
                <D:getlastmodified>Sat, 13 Apr 2019 16:52:08 GMT</D:getlastmodified>
                <D:getcontenttype>text/plain; charset=utf-8</D:getcontenttype>
                <D:getetag>"159516cfe790beecd"</D:getetag>
                <D:supportedlock>
                    <D:lockentry xmlns:D="DAV:">
                        <D:lockscope>
                            <D:exclusive/>
                        </D:lockscope>
                        <D:locktype>
                            <D:write/>
                        </D:locktype>
                    </D:lockentry>
                </D:supportedlock>
            </D:prop>
            <D:status>HTTP/1.1 200 OK</D:status>
        </D:propstat>
    </D:response>
</D:multistatus>
```