#include <thread>
#include <mutex>
#include <iostream>

int main() {
    std::mutex m;
    std::thread t([](){
        std::cout << "Thread works!" << std::endl;
    });
    t.join();
    return 0;
}
