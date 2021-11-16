/*
 * This file contains an example in Python for an AI controlled client.
 * Use this example to program your own AI in Python.
 */

import java.io.*;
import java.net.Socket;

public class CloudWar {

    // CONFIG
    public static final String host = "localhost";
    public static final int port = 3333;
    public static final String name = "HansAI";
    public static final String color = "orange";

    // inner vars
    private final Socket socket;
    private final BufferedReader reader;
    private final BufferedWriter writer;

    // constructor
    public CloudWar() throws IOException {
        // TCP connection
        this.socket = new Socket(host, port);

        // reader & writer
        this.reader = new BufferedReader(new InputStreamReader(this.socket.getInputStream()));
        this.writer = new BufferedWriter(new OutputStreamWriter(this.socket.getOutputStream()));
    }

    public static void main(String[] args) throws IOException, InterruptedException {
       CloudWar cw = new CloudWar();
       cw.myAI();
    }

    // implement your AI
    public void myAI() throws IOException, InterruptedException {

        // set name, color and start the game
        name(CloudWar.name);
        color(CloudWar.color);
        play();

        // get world status
        String json = stat();
        System.out.println(json);

        /*
            // parse some world stats
            world['Width']  // game board size (DEFAULT: 2048)
            world['Height']  // game board size (DEFAULT: 1152)
            world['GameSpeed']  // updates per second (DEFAULT: 60)
            world['Iteration']  // increases with every server update
            world['WorldVapor']  // vapor of all clouds together
            world['Alive']  // active clouds
            world['Clouds']  // cloud list

            // cloud list
            world['Clouds'][n]['Player']  // only player controlled clouds have names
            world['Clouds'][n]['Color']  // cloud color
            world['Clouds'][n]['Vapor']  // cloud vapor (mass)
            world['Clouds'][n]['Pos']['X']  // x position
            world['Clouds'][n]['Pos']['Y']  // y position
            world['Clouds'][n]['Vel']['X']  // x velocity (speed)
            world['Clouds'][n]['Vel']['Y']  // y velocity (speed)
       */

        // move around
        Thread.sleep(2000);
        moveByAngle(0, 10);  // move left
        Thread.sleep(2000);
        moveByAngle(90, 10);  // move up
        Thread.sleep(2000);
        moveByAngle(180, 10);  // move right
        Thread.sleep(2000);
        moveByAngle(270, 10);  // move down
        Thread.sleep(2000);

        // it makes no sense
        kill();
        close();
    }

    // ------ Helper ------------------------------------------------------------------------------------------ //

    private String cmdWriteRead(String cmd) throws IOException {

        // remove protocol break
        cmd = cmd.replace("\n", "");
        cmd = cmd.replace("\r", "");

        // send command
        this.writer.write(cmd);
        this.writer.newLine();
        this.writer.flush();
        System.out.println("SEND: " + cmd);

        // read response
        String resp = this.reader.readLine();
        System.out.println("RESP: " + resp);

        // return
        return resp;
    }

    // ------ Commands ---------------------------------------------------------------------------------------- //

    // Close disconnects from the server.
    // The controlled cloud remains unchanged (use Kill() before this call).
    // Returns the server response (OK or ERR) as a string.
    private String close() throws IOException {
        String resp = cmdWriteRead("quit");
        this.socket.close();
        return resp;
    }

    // Stat returns the world status as a json string.
    // Use core.World FromJson() to parse the string.
    private String stat() throws IOException {
        return cmdWriteRead("list");
    }

    // Name set the player name
    // Use this before calling Play()
    // Returns the server response (OK or ERR) as a string.
    private String name(String n) throws IOException {
        String cmd = "name" + n;
        return cmdWriteRead(cmd);
    }

    // Color set the player color.
    // 'blue', 'gray', 'orange', 'purple' or 'red'
    // Use this before calling Play()
    // Returns the server response (OK or ERR) as a string.
    private String color(String c) throws IOException {
        String cmd = "type" + c;
        return cmdWriteRead(cmd);
    }

    // Play creates a new player cloud.
    // The attributes of Name() and Color() are used.
    // Returns the server response (OK or ERR) as a string.
    private String play() throws IOException {
        return cmdWriteRead("play");
    }

    // Move sends a move command for your player cloud to the server.
    // Returns the server response (OK or ERR) as a string.
    private String move(int x, int y) throws IOException {
        String cmd = "move" + x + ";" + y;
        return cmdWriteRead(cmd);
    }

    // MoveByAngle sends a move command for your player cloud to the server.
    // Returns the server response (OK or ERR) as a string.
    private String moveByAngle(float angle, float strength) throws IOException {
        int x = (int) (Math.cos(Math.PI / 180 * angle) * strength * (-1));
        int y = (int) (Math.sin(Math.PI / 180 * angle) * strength * (-1));
        return move(x, y);
    }

    // Kill blasts the controlled cloud and removes it from the game.
    // Returns the server response (OK or ERR) as a string.
    private String kill() throws IOException {
        return cmdWriteRead("kill");
    }
}
