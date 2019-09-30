package config

func LogDefautConfig() string {
	return `
        <seelog type="sync">
        	<outputs formatid="main">
                <console />
        	</outputs>
            <formats>
                <format id="main" format="%Date %Time %File:%Line [%Level] %Msg%n"/>
            </formats>
        </seelog>
    `
}
