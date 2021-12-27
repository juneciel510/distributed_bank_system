import React from "react";
import styled from "styled-components";
// import { Route, Switch as Switching } from "react-router";
import {
  BrowserRouter,
  Routes as Switching,
  Route,
  Link,
} from "react-router-dom";
import Navbar from "./components/Navbar";
import Service from "./pages/Service";
import Reconfiguration from "./pages/Reconfiguration";
//import "./App.css";

var client = new WebSocket("ws://127.0.0.1:12110/ws");
class App extends React.Component {
  constructor() {
    super();
    //initialize the clientID, clientSeq and accountNum
    this.state = {
      clientID: "123",
      clientSeq: 1,
      command: null,
      accountNum: 42,
      op: 0,
      amount: null,
      ws: null,
      message: [],
      servers: [
        "127.0.0.1:12110",
        "127.0.0.1:12111",
        "127.0.0.1:12112",
        "127.0.0.1:12113",
        "127.0.0.1:12114",
        "127.0.0.1:12115",
        "127.0.0.1:12116",
      ],
      index: 0,
      length: null,
      serverIDs: [1, 2, 3, 4, 5, 6, 7],
      checkedState: [false, false, false, false, false, false, false],
    };
    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
    //this.opNumberToString = this.opNumberToString.bind(this);
    this.handleChange_CB = this.handleChange_CB.bind(this);
    this.handleSubmit_CB = this.handleSubmit_CB.bind(this);
    this.state.length = this.state.servers.length;
  }
  opNumberToString(op) {
    switch (Number(op)) {
      case 0:
        return "Balance";
      case 1:
        return "Deposit";
      case 2:
        return "Withdrawl";
    }
  }
  createOpMessage(op) {
    switch (Number(op)) {
      case 0:
        return `Account number: ${this.state.accountNum}
        Operation:Balance
      `;
      case 1:
        return `Account number: ${this.state.accountNum}
        Operation:Deposit
        Amount: ${this.state.amount}
      `;
      case 2:
        return `Account number: ${this.state.accountNum}
        Operation:Withdrawl
        Amount: ${this.state.amount}
      `;
    }
  }
  //establish the websoket connection

  handleChange(event) {
    this.setState({ [event.target.name]: event.target.value }); //handle multiple event.target.value, store in an array

    console.log(this.state[event.target.name]);
  }

  handleChange_CB(event) {
    const updatedCheckedState = this.state.checkedState.map((item, index) =>
      index === Number(event.target.name) ? !item : item
    );
    this.setState({
      checkedState: updatedCheckedState,
    });
  }

  handleSubmit(event) {
    let opString = this.opNumberToString(this.state.op);
    if (this.state.op == 0) {
      alert(
        "Please check the transction you are going to submit.\nThe operation: " +
          opString +
          "\nAccount number: " +
          this.state.accountNum
      );
    } else {
      alert(
        "Please check the transction you are going to submit.\nThe operation: " +
          opString +
          "\nAccount number: " +
          this.state.accountNum +
          "\nAmount: " +
          this.state.amount
      );
    }

    let opMessage = `Account number: ${this.state.accountNum}
      Operation:${opString}
      Amount:${this.state.amount}
    `;
    //resSting = "Account number:" + accountNum + "\nBalance:" + balance;

    opMessage = this.createOpMessage(this.state.op);
    console.log(opMessage);

    this.setState({ message: this.state.message.concat(opMessage) });

    //console.log("typeof(this.state.amount)", typeof this.state.amount);
    let op = Number(this.state.op);
    let accountNum = Number(this.state.accountNum);
    let amount = Number(this.state.amount);
    console.log("typeof(amount)", typeof amount, amount);
    let msg = {
      Command: "VALUE",
      Parameter: {
        ClientID: this.state.clientID,
        ClientSeq: this.state.clientSeq,
        Noop: false,
        AccountNum: accountNum,
        Txn: {
          Op: op,
          Amount: amount,
        },
      },
    };

    this.state.ws.send(JSON.stringify(msg));
    this.setState({ clientSeq: this.state.clientSeq + 1 });
    console.log("this.state.clientSeq", this.state.clientSeq);
    event.preventDefault();
  }

  handleSubmit_CB(event) {
    console.log("this.state.checkedState", this.state.checkedState);
    const serverChosen = [];
    this.state.checkedState.map((item, index) =>
      item === true ? serverChosen.push(this.state.serverIDs[index]) : null
    );
    console.log("serverChosen", serverChosen);
    let msg = {
      Command: "RECONFIG",
      Parameter: {
        Timestamp: new Date().getTime(),
        NewerServStr: serverChosen.join(),
      },
    };
    console.log("msg", msg);
    this.state.ws.send(JSON.stringify(msg));
    event.preventDefault();
  }

  openEventListener = (event) => {
    //send client message
    this.setState({ ws: client });
    console.log("Websocket Client Connected: ", client.url);
    let msg = {
      Command: "CLIENT",
      Parameter: {
        ClientID: this.state.clientID,
      },
    };
    this.state.ws.send(JSON.stringify(msg));
  };

  incomingMessageListener = (message) => {
    let messageData = JSON.parse(message.data);
    console.log(messageData);

    switch (messageData["Command"]) {
      case "LEADER":
        // console.log(messageData["Parameter"]);
        // console.log(
        //   "typeof(messageData[Parameter])",
        //   typeof messageData["Parameter"]
        // );
        client = new WebSocket("ws://" + messageData["Parameter"] + "/ws");
        this.setState({
          ws: new WebSocket("ws://" + messageData["Parameter"] + "/ws"),
        });
        console.log("this.state.ws.url", this.state.ws.url);
        this.wsMount();
        break;
      case "RESPONSE":
        console.log("receive response", messageData["Parameter"]["TxnRes"]);
        let accountNum = messageData["Parameter"]["TxnRes"]["AccountNum"];
        let balance = messageData["Parameter"]["TxnRes"]["Balance"];
        let resString = `Account number: ${accountNum}
        Balance:${balance}
        `;
        //resSting = "Account number:" + accountNum + "\nBalance:" + balance;
        console.log(resString);
        message = this.state.message.concat(resString);
        this.setState({ message: message });
        break;
      default:
      //console.log(messageData);
    }
    //todo:message dispatching
  };

  closeSocket = (event) => {
    console.log("You are disconnected");
    //reconnect
    //give up after a cetain times of attempts of re-connention
    if (this.state.index < 8) {
      setTimeout(() => {
        console.log("Retrying connection times:", this.state.index);
        client = new WebSocket(
          "ws://" +
            this.state.servers[this.state.index % this.state.length] +
            "/ws"
        );
        this.setState({ index: this.state.index + 1 });
        client.addEventListener("open", this.openEventListener);
        client.addEventListener("message", this.incomingMessageListener);
        client.addEventListener("close", this.closeSocket);
      }, 5000);
    }
  };

  wsMount() {
    console.log("wsMount()");
    /*
    this.state.ws.addEventListener("open", this.openEventListener);
    this.state.ws.addEventListener("message", this.incomingMessageListener);
    this.state.ws.addEventListener("close", this.closeSocket);
    */
    client.addEventListener("open", this.openEventListener);
    client.addEventListener("message", this.incomingMessageListener);
    client.addEventListener("close", this.closeSocket);
  }

  componentDidMount() {
    console.log("componentDidMount()");
    //this.newWS();
    //console.log("client.readyState", this.state.ws.readyState);
    //console.log("client:", this.state.ws);
    this.wsMount();
  }

  render() {
    const messages = this.state.message.map((msg, index) => (
      <p key={index}>{msg}</p>
    ));
    return (
      <div className="App">
        <Navbar></Navbar>
        <BrowserRouter>
          <Switching>
            <Route
              path="service"
              element={
                <Service
                  handleSubmit={this.handleSubmit}
                  op={this.state.op}
                  handleChange={this.handleChange}
                  accountNum={this.state.accountNum}
                  amount={this.state.amount}
                  messages={messages}
                />
              }
            />
            <Route
              path="reconfiguration"
              element={
                <Reconfiguration
                  handleSubmit_CB={this.handleSubmit_CB}
                  handleChange_CB={this.handleChange_CB}
                  serverIDs={this.state.serverIDs}
                  checkedState={this.state.checkedState}
                />
              }
            />
          </Switching>
        </BrowserRouter>
      </div>
    );
  }
}

const AppStyled = styled.div`
  position: absolute;
  left: 35%;
  .row {
    margin-top: 5px;
    select {
      margin-left: 5px;
    }
    .amount {
      margin-left: 5px;
    }
  }
`;
export default App;
