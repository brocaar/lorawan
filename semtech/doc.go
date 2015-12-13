/*

Package semtech implements the protocol written by Semtech for communication
between the gateway and the server.

The following upstream packet types are implemented:

    * PUSH_DATA
    * PUSH_ACK

The following downstream packet types are implemented:

    * PULL_DATA
    * PULL_ACK
    * PULL_RESP

The specification can be found at:
https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT

*/
package semtech
