#include <iostream>
#include <boost/asio.hpp>
#include <boost/json/src.hpp>

namespace asio = boost::asio;
namespace ip = asio::ip;

constexpr
auto tcp_ip{"127.0.0.1"}, tcp_port{"3333"};


int main() try {
	asio::io_context io_context;
	ip::tcp::socket socket{io_context};
	ip::tcp::resolver resolver{io_context};

	auto send = [&](std::string msg) {
		msg += "\n";
		asio::write(socket, asio::const_buffer(msg.data(), msg.size()));
		msg.clear();
		asio::read_until(socket, asio::dynamic_buffer(msg), '\n');
		if(!msg.starts_with('{')) std::cout << msg << std::flush;
		return msg;
	};

	asio::connect(socket, resolver.resolve(tcp_ip, tcp_port));

	//register remote player
	send("nameMike");
	send("typered");
	send("play");

	//get current game state
	auto json{boost::json::parse(send("list"))};
	std::cout << json.at("Height").as_int64() << "\n";


	//send move to server
	send("move10;10");
} catch(const std::exception & exc) {
	std::cerr << "ERROR: " << exc.what() << "\n";
}
