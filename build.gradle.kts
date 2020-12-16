val loadbalanceServerImage = System.getenv("REPO") + version
val USER = System.getenv("USER")
val KEY = System.getenv("KEY")

tasks {

    register("lbbuild") {
        group = "Build"

        doLast {
           exec {
                commandLine("bash", "-c", "./generate.sh")
           }
            exec {
                commandLine("docker build --no-cache --build-arg AUSER=$USER --build-arg AKEY=$KEY . -f ./Dockerfile -t $loadbalanceServerImage".split(" "))
            }

        }
    }

    register("lbpublish") {
        group = "Build"

        doLast {
            exec {
                commandLine("docker push $loadbalanceServerImage".split(" "))
            }
        }
    }
}
