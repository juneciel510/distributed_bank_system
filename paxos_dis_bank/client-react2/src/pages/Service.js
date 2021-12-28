import React from "react";
import styled from "styled-components";

class Service extends React.Component {
  render() {
    return (
      <div className="Service">
        <form onSubmit={this.props.handleSubmit}>
          <div className="row">
            <label>
              Choose your transction operation:
              <select
                className="select"
                name="op" //name is same as the state name
                value={this.props.op}
                onChange={this.props.handleChange}
              >
                <option value="0">Balance</option>
                <option value="1">Deposit</option>
                <option value="2">Withdrawal</option>
              </select>
            </label>
            <br />
          </div>
          <div className="row">
            <label>
              Select the account number:
              <select
                className="select"
                name="accountNum" //name is same as the state name
                value={this.props.accountNum}
                onChange={this.props.handleChange}
              >
                <option value="42">42</option>
                <option value="52">52</option>
                <option value="62">62</option>
              </select>
            </label>
            <br />
          </div>
          <div className="row">
            <label>
              Please input your amount:
              <input
                className="amount"
                name="amount"
                type="number"
                value={this.props.amount}
                onChange={this.props.handleChange}
              />
            </label>
            <div className="row">
              <input className="submit" type="submit" value="Submit" />
              <br />
            </div>
          </div>
        </form>
        {this.props.messages}
      </div>
    );
  }
}

export default Service;
