// *****************************************************
// AUTHOR  : Alex Stocks
// EMAIL   : alexstocks@foxmail.com
// MOD     : 2016-10-26 11:16
// FILE    : HessianServer.java
// ******************************************************

package com.ikurento.hessian;

import org.eclipse.jetty.server.NCSARequestLog;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.handler.HandlerList;
import org.eclipse.jetty.server.handler.RequestLogHandler;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHolder;

public class HessianServer {
    private static final boolean debug = false;

    public static void main(String[] args) throws Exception {
        Server server = new Server(8000);

        HandlerList handlers = new HandlerList();

        ServletContextHandler context = new ServletContextHandler(1);

        // create echo servlet
        ServletHolder echoSH = new ServletHolder(new EchoImpl());
        echoSH.setInitParameter("debug", "true");

        // create math servlet
        ServletHolder mathSH = new ServletHolder(new MathImpl());
        mathSH.setInitParameter("debug", "true");

        // create datatype servlet
        ServletHolder dtSH = new ServletHolder(new DataTypeImpl());
        dtSH.setInitParameter("debug", "true");

        // add servlet
        context.addServlet(echoSH, "/echo");
        context.addServlet(mathSH, "/math");
        context.addServlet(dtSH, "/datatype");

        if (debug) {
            NCSARequestLog requestLog = new NCSARequestLog();
            requestLog.setFilename("./logs/yyyy_mm_dd.request.log");
            requestLog.setFilenameDateFormat("yyyy_MM_dd");
            requestLog.setRetainDays(90);
            requestLog.setAppend(true);
            requestLog.setExtended(true);
            requestLog.setLogCookies(false);
            requestLog.setLogTimeZone("GMT");

            RequestLogHandler requestLogHandler = new RequestLogHandler();
            requestLogHandler.setRequestLog(requestLog);
            handlers.addHandler(requestLogHandler);
        }

        // add handler
        handlers.addHandler(context);

        // set handler
        server.setHandler(handlers);
        server.start();
        server.join();
    }
}
